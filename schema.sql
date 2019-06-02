CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.modified = now();
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TYPE TASK_STATUS AS ENUM ('created', 'in_progress', 'completed', 'error', 'timeout');

CREATE TABLE tasks (
  task_id BIGSERIAL PRIMARY KEY,
  created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  modified TIMESTAMP DEFAULT NULL,
  name TEXT,
  pipe TEXT[],
  status TASK_STATUS DEFAULT 'created',
  attempts int DEFAULT 0,
  max_attempts int DEFAULT 1,
  backoff timestamp DEFAULT CURRENT_TIMESTAMP,
  payload JSONB,
  result JSONB[]
);

CREATE TRIGGER update_task_modtime
BEFORE UPDATE ON tasks
FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
