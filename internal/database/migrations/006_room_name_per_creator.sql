-- #20: 将房间名从全局唯一改为 creator_id + name 联合唯一。
--
-- 原约束 (rooms_name_key) 要求所有房间名全局不重复，随着用户增多好名字
-- 会被占完，且不同用户的私密房间也无法同名。
-- 新策略：同一创建者不能创建同名房间，但不同创建者可以同名。
--
-- 注意：这是不可逆迁移（原 UNIQUE 索引已删除）。如需回滚，请手动重建：
--   CREATE UNIQUE INDEX rooms_name_key ON rooms(name);

ALTER TABLE rooms DROP CONSTRAINT IF EXISTS rooms_name_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_rooms_name_per_creator
    ON rooms (creator_id, name);
