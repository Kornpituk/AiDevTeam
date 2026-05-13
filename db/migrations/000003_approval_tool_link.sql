ALTER TABLE human_approvals ADD COLUMN tool_call_id UUID REFERENCES agent_tool_calls(id);
CREATE INDEX IF NOT EXISTS idx_human_approvals_tool_call_id ON human_approvals(tool_call_id);
CREATE INDEX IF NOT EXISTS idx_human_approvals_status_type ON human_approvals(status, approval_type);
