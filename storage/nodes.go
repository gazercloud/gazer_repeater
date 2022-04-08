package storage

import (
	"errors"
	"github.com/gazercloud/gazer_repeater/logger"
	"github.com/jackc/pgx"
	"strconv"
)

type NodeResponseItem struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	CurrentRepeater string `json:"current_repeater"`
}

type NodesResponse struct {
	Items []NodeResponseItem `json:"items"`
}

func (c *Storage) Nodes(userId int64) (res *NodesResponse, err error) {
	res = &NodesResponse{}
	res.Items = make([]NodeResponseItem, 0)

	err = c.checkConnection()
	if err != nil {
		return
	}

	var rows *pgx.Rows
	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return nil, err
	}

	rows, err = c.db.Query("SELECT id, name FROM nodes WHERE user_id=$1", userId)
	if err != nil {
		_ = tr.Rollback()
		return nil, err
	}
	for rows.Next() {
		item := NodeResponseItem{}
		var itemIsInt int64
		err = rows.Scan(&itemIsInt, &item.Name)
		if err != nil {
			break
		}
		item.Id = strconv.FormatInt(itemIsInt, 10)
		res.Items = append(res.Items, item)
	}
	rows.Close()
	_ = tr.Rollback()

	return
}

func (c *Storage) NodeAdd(userId int64, name string) (nodeId string, err error) {
	err = c.checkConnection()
	if err != nil {
		return
	}

	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return
	}

	logger.Println("Storage NodeAdd", userId, name)

	// Max node count for user
	var maxNodesPerUser int64
	{
		var res *pgx.Rows
		res, err = tr.Query("SELECT (max_nodes_count + free_nodes) as max_nodes_count FROM users WHERE id=$1", userId)
		if err != nil {
			_ = tr.Rollback()
			return
		}
		if res.Next() {
			err = res.Scan(&maxNodesPerUser)
			logger.Println("Storage NodeAdd maxNodesPerUser", maxNodesPerUser)
			if err != nil {
				res.Close()
				_ = tr.Rollback()
				return
			}
		}
		res.Close()
	}

	// nodes count for user
	var currentNodesCountForUser int64
	{
		var res *pgx.Rows
		res, err = tr.Query("SELECT count(*) FROM nodes WHERE user_id=$1", userId)
		if err != nil {
			_ = tr.Rollback()
			return
		}
		if res.Next() {
			err = res.Scan(&currentNodesCountForUser)
			logger.Println("Storage NodeAdd currentNodesCountForUser", currentNodesCountForUser)
			if err != nil {
				res.Close()
				_ = tr.Rollback()
				return
			}
		}
		res.Close()
	}

	if currentNodesCountForUser >= maxNodesPerUser {
		err = errors.New("your tariff plan does not allow you to create more than " + strconv.FormatInt(maxNodesPerUser, 10) + " nodes")
		_ = tr.Rollback()
		return
	}

	// Create new node
	{
		var res *pgx.Rows
		res, err = tr.Query("INSERT INTO nodes (id, name, user_id) VALUES (nextval('seq_node_id'), $1, $2) RETURNING id", name, userId)
		if err != nil {
			_ = tr.Rollback()
			return
		}
		if res.Next() {
			var nodeIdInt int64
			err = res.Scan(&nodeIdInt)
			logger.Println("Storage NodeAdd nodeId", nodeId)
			if err != nil {
				res.Close()
				_ = tr.Rollback()
				return
			}
			nodeId = strconv.FormatInt(nodeIdInt, 10)
		} else {
			err = errors.New("no node id returned")
			logger.Println("Storage NodeAdd NO nodeId ERROR")
		}
		res.Close()
	}

	err = tr.Commit()

	return
}

func (c *Storage) NodeUpdate(userId int64, nodeId string, name string) (err error) {
	var nodeIdInt int64
	nodeIdInt, err = strconv.ParseInt(nodeId, 10, 64)
	if err != nil {
		return
	}

	err = c.checkConnection()
	if err != nil {
		return
	}

	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return
	}

	_, err = tr.Exec("UPDATE nodes SET name=$1 WHERE id=$2 AND user_id=$3", name, nodeIdInt, userId)
	if err != nil {
		_ = tr.Rollback()
		return
	}
	_ = tr.Commit()

	return
}

func (c *Storage) NodeRemove(userId int64, nodeId string) (err error) {
	var nodeIdInt int64
	nodeIdInt, err = strconv.ParseInt(nodeId, 10, 64)
	if err != nil {
		return
	}

	err = c.checkConnection()
	if err != nil {
		return
	}

	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return
	}

	_, err = tr.Exec("DELETE FROM nodes WHERE id=$1 AND user_id=$2", nodeIdInt, userId)
	if err != nil {
		_ = tr.Rollback()
		return
	}
	_ = tr.Commit()

	return
}
