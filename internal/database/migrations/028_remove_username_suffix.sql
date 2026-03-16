-- 028_remove_username_suffix.sql: Remove '#xxxx' suffix from usernames and handle collisions
DO $$
DECLARE 
    r RECORD;
    new_uname TEXT;
    base_uname TEXT;
    counter INT;
    final_uname VARCHAR(32);
BEGIN
    FOR r IN 
        SELECT id, username FROM users WHERE username ~ '#[0-9]{4}$'
    LOOP
        -- 切割剔除后缀，例如 张三#0001 变为 张三
        base_uname := regexp_replace(r.username, '#[0-9]{4}$', '');
        
        -- 防御性逻辑：如果原先只有 '#0001'，裁剪后为空，给个保底。如果本身就是 user，后续防重名会自动处理
        IF base_uname = '' THEN
            base_uname := 'user';
        END IF;

        new_uname := base_uname;
        counter := 1;
        final_uname := new_uname;
        
        -- 重名检测与自增处理
        WHILE EXISTS(SELECT 1 FROM users WHERE username = final_uname AND id <> r.id) LOOP
            -- 如果名字过长，砍掉后几位再拼数字
            IF length(base_uname || counter::text) > 32 THEN
                final_uname := substring(base_uname, 1, 32 - length(counter::text)) || counter::text;
            ELSE
                final_uname := base_uname || counter::text;
            END IF;
            counter := counter + 1;
        END LOOP;

        -- 执行更新
        UPDATE users 
        SET 
            username = final_uname,
            updated_at = NOW()
        WHERE id = r.id;

        -- 全局修复带 @ 标记的历史文本引用 (包括群组聊天和私聊)
        UPDATE messages 
        SET content = replace(content, '@' || r.username, '@' || final_uname)
        WHERE content LIKE '%@' || r.username || '%';

        UPDATE direct_messages 
        SET content = replace(content, '@' || r.username, '@' || final_uname)
        WHERE content LIKE '%@' || r.username || '%';

    END LOOP;
END $$;
