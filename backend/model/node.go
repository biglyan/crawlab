package model

import (
	"crawlab/constants"
	"crawlab/database"
	"crawlab/services/register"
	"github.com/apex/log"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"runtime/debug"
	"time"
)

type Node struct {
	Id          bson.ObjectId `json:"_id" bson:"_id"`
	Name        string        `json:"name" bson:"name"`
	Status      string        `json:"status" bson:"status"`
	Ip          string        `json:"ip" bson:"ip"`
	Port        string        `json:"port" bson:"port"`
	Mac         string        `json:"mac" bson:"mac"`
	Description string        `json:"description" bson:"description"`
	// 用于唯一标识节点，可能是mac地址，可能是ip地址
	Key string `json:"key" bson:"key"`

	// 前端展示
	IsMaster bool `json:"is_master"`

	UpdateTs     time.Time `json:"update_ts" bson:"update_ts"`
	CreateTs     time.Time `json:"create_ts" bson:"create_ts"`
	UpdateTsUnix int64     `json:"update_ts_unix" bson:"update_ts_unix"`
}

func (n *Node) Save() error {
	s, c := database.GetCol("nodes")
	defer s.Close()
	n.UpdateTs = time.Now()
	if err := c.UpdateId(n.Id, n); err != nil {
		return err
	}
	return nil
}

func (n *Node) Add() error {
	s, c := database.GetCol("nodes")
	defer s.Close()
	n.Id = bson.NewObjectId()
	n.UpdateTs = time.Now()
	n.UpdateTsUnix = time.Now().Unix()
	n.CreateTs = time.Now()
	if err := c.Insert(&n); err != nil {
		debug.PrintStack()
		return err
	}
	return nil
}

func (n *Node) Delete() error {
	s, c := database.GetCol("nodes")
	defer s.Close()
	if err := c.RemoveId(n.Id); err != nil {
		debug.PrintStack()
		return err
	}
	return nil
}

func (n *Node) GetTasks() ([]Task, error) {
	tasks, err := GetTaskList(bson.M{"node_id": n.Id}, 0, 10, "-create_ts")
	//tasks, err := GetTaskList(nil, 0, 10, "-create_ts")
	if err != nil {
		debug.PrintStack()
		return []Task{}, err
	}

	return tasks, nil
}

func GetNodeList(filter interface{}) ([]Node, error) {
	s, c := database.GetCol("nodes")
	defer s.Close()

	var results []Node
	if err := c.Find(filter).All(&results); err != nil {
		log.Error("get node list error: " + err.Error())
		debug.PrintStack()
		return results, err
	}
	return results, nil
}

func GetNode(id bson.ObjectId) (Node, error) {
	s, c := database.GetCol("nodes")
	defer s.Close()

	var node Node
	if err := c.FindId(id).One(&node); err != nil {
		if err != mgo.ErrNotFound {
			log.Errorf(err.Error())
			debug.PrintStack()
		}
		return node, err
	}
	return node, nil
}

func GetNodeByKey(key string) (Node, error) {
	s, c := database.GetCol("nodes")
	defer s.Close()

	var node Node
	if err := c.Find(bson.M{"key": key}).One(&node); err != nil {
		if err != mgo.ErrNotFound {
			log.Errorf(err.Error())
			debug.PrintStack()
		}
		return node, err
	}
	return node, nil
}

func UpdateNode(id bson.ObjectId, item Node) error {
	s, c := database.GetCol("nodes")
	defer s.Close()

	var node Node
	if err := c.FindId(id).One(&node); err != nil {
		return err
	}

	if err := item.Save(); err != nil {
		return err
	}
	return nil
}

func GetNodeTaskList(id bson.ObjectId) ([]Task, error) {
	node, err := GetNode(id)
	if err != nil {
		return []Task{}, err
	}
	tasks, err := node.GetTasks()
	if err != nil {
		return []Task{}, err
	}
	return tasks, nil
}

func GetNodeCount(query interface{}) (int, error) {
	s, c := database.GetCol("nodes")
	defer s.Close()

	count, err := c.Find(query).Count()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// 节点基本信息
func GetNodeBaseInfo() (ip string, mac string, key string, error error) {
	ip, err := register.GetRegister().GetIp()
	if err != nil {
		debug.PrintStack()
		return "", "", "", err
	}

	mac, err = register.GetRegister().GetMac()
	if err != nil {
		debug.PrintStack()
		return "", "", "", err
	}

	key, err = register.GetRegister().GetKey()
	if err != nil {
		debug.PrintStack()
		return "", "", "", err
	}
	return ip, mac, key, nil
}

// 根据redis的key值，重置node节点为offline
func ResetNodeStatusToOffline(list []string) {
	nodes, _ := GetNodeList(nil)
	for _, node := range nodes {
		hasNode := false
		for _, key := range list {
			if key == node.Key {
				hasNode = true
				break
			}
		}
		if !hasNode || node.Status == "" {
			node.Status = constants.StatusOffline
			if err := node.Save(); err != nil {
				log.Errorf(err.Error())
				return
			}
			continue
		}
	}
}
