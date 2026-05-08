CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS ai_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    plan TEXT,
    review_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ai_task_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES ai_tasks(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_task_events_task_id ON ai_task_events(task_id);

CREATE TABLE IF NOT EXISTS ai_task_artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES ai_tasks(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    artifact_type VARCHAR(50) NOT NULL,
    content TEXT,
    content_type VARCHAR(100),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_task_artifacts_task_id ON ai_task_artifacts(task_id);
