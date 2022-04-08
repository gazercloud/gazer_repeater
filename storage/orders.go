package storage

import (
	"errors"
	"github.com/gazercloud/gazer_repeater/logger"
	"github.com/jackc/pgx"
)

func (c *Storage) AddOrder(userId int64, name string, key string, active bool, product string, quantity int64, price float64, rawData string) (orderId int64, err error) {
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
	var countOfOrders int64
	{
		var res *pgx.Rows
		res, err = tr.Query("SELECT count(*) FROM orders WHERE name=$1", name)
		if err != nil {
			_ = tr.Rollback()
			return
		}
		if res.Next() {
			err = res.Scan(&countOfOrders)
			logger.Println("Storage Add Order. Count:", countOfOrders)
			if err != nil {
				res.Close()
				_ = tr.Rollback()
				return
			}
		}
		res.Close()
	}

	if countOfOrders > 0 {
		err = errors.New("order exists: " + name)
		_ = tr.Rollback()
		return
	}

	// Create new order
	{
		var res *pgx.Rows
		res, err = tr.Query("INSERT INTO orders (id, user_id, name, key, active, product, quantity, price, rawdata) VALUES (nextval('seq_order_id'), $1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
			userId, name, key, active, product, quantity, price, rawData)
		if err != nil {
			_ = tr.Rollback()
			return
		}
		if res.Next() {
			var orderIdInt int64
			err = res.Scan(&orderIdInt)
			logger.Println("Storage NodeAdd nodeId", orderIdInt)
			if err != nil {
				res.Close()
				_ = tr.Rollback()
				return
			}
			orderId = orderIdInt
		} else {
			err = errors.New("no order id returned")
			logger.Println("Storage AddOrder NO orderId ERROR")
		}
		res.Close()
	}

	err = tr.Commit()

	return
}

func (c *Storage) UpdateMaxNodesCount(userId int64) (err error) {
	err = c.checkConnection()
	if err != nil {
		return
	}

	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return
	}

	logger.Println("Storage UpdateMaxNodesCount", userId)

	maxNodesCount := int64(0)
	{
		var res *pgx.Rows
		res, err = tr.Query("SELECT product, quantity FROM orders WHERE user_id=$1 AND active = TRUE", userId)
		if err != nil {
			_ = tr.Rollback()
			return
		}
		for res.Next() {
			var product string
			var quantity int64
			err = res.Scan(&product, &quantity)
			if err != nil {
				res.Close()
				_ = tr.Rollback()
				return
			}

			if product == "gazercloud-pro" {
				maxNodesCount += quantity
			}
		}
		res.Close()
	}

	// Update maxNodesCount
	{
		logger.Println("Storage UpdateMaxNodesCount userId:", userId, "newValue:", maxNodesCount)
		_, err = tr.Exec("UPDATE users SET max_nodes_count = $1 WHERE id = $2",
			maxNodesCount, userId)
		if err != nil {
			_ = tr.Rollback()
			return
		}
	}

	err = tr.Commit()

	return
}
