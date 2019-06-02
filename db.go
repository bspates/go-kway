package main

import (
	"encoding/json"
	//"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const STATUS_CREATED string = "created"
const STATUS_ERROR string = "error"
const STATUS_COMPLETED string = "completed"
const STATUS_IN_PROGRESS string = "in_progress"

func result(db *sqlx.DB, tasks []*Task) error {
	jsonTasks, err := json.Marshal(tasks)
	jsonStr := string(jsonTasks)

	if err != nil {
		return err
	}
	query := `
WITH updates AS (
	SELECT 
		m->>'task_id' AS task_id, 
    m->>'status' AS status,
		m->>'result' AS result
	FROM json_array_elements($1::JSON) AS m 	
), doit AS (
UPDATE tasks t
SET status = u.status::TASK_STATUS,
	result = array_append(t.result, u.result::JSONB)
	backoff = CURRENT_TIMESTAMP + INTERVAL '2 seconds' ^ t.attempts 
FROM updates u 
WHERE u.task_id::BIGINT = t.task_id
) SELECT task_id, status, result FROM updates;`

	db.QueryRow(query, jsonStr)
	return nil
}

func dequeue(db *sqlx.DB, concurrency int) ([]*Task, error) {
	queryTemplate := `
WITH selected AS (
	SELECT 
		task_id,
		created,
		modified,
		name,
		pipe,
		status,
		attempts,
		max_attempts,
		backoff,
		payload
	FROM tasks 
	WHERE (
		status = $1::TASK_STATUS 
		OR status = $2::TASK_STATUS
	)
	AND backoff < CURRENT_TIMESTAMP
	AND max_attempts > attempts
	FOR UPDATE SKIP LOCKED
	LIMIT $3::INT
), update_selected AS (
	UPDATE tasks t
	SET status = $4::TASK_STATUS,
		attempts = t.attempts + 1
	FROM selected s
	WHERE t.task_id = s.task_id
) SELECT * FROM selected;`

	statement, err := db.Prepare(queryTemplate)
	if err != nil {
		return nil, err
	}

	rows, err := statement.Query(STATUS_CREATED, STATUS_ERROR, concurrency, STATUS_IN_PROGRESS)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*Task{}
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.TaskId,
			&task.Created,
			&task.Modified,
			&task.Name,
			pq.Array(&task.Pipe),
			&task.Status,
			&task.Attempts,
			&task.MaxAttempts,
			&task.Backoff,
			&task.Payload)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}
	return tasks, nil
}
