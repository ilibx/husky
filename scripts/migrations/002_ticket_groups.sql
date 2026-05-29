-- 工单-飞书群绑定表
CREATE TABLE IF NOT EXISTS ticket_groups (
    id SERIAL PRIMARY KEY,
    ticket_id INTEGER NOT NULL UNIQUE REFERENCES tickets(id) ON DELETE CASCADE,
    group_id VARCHAR(255) NOT NULL,
    group_name VARCHAR(255),
    join_link TEXT,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_ticket_groups_ticket_id ON ticket_groups(ticket_id);
CREATE INDEX idx_ticket_groups_group_id ON ticket_groups(group_id);
CREATE INDEX idx_ticket_groups_status ON ticket_groups(status);
