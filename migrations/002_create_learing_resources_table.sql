-- Create learning_resources table
CREATE TABLE IF NOT EXISTS learning_resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(500) NOT NULL,
    url TEXT,
    description TEXT,
    technology VARCHAR(100) NOT NULL,
    type VARCHAR(20) CHECK (type IN ('article', 'video', 'course', 'documentation', 'tutorial', 'book', 'podcast')) NOT NULL,
    status VARCHAR(20) CHECK (status IN ('to-read', 'reading', 'completed', 'bookmarked')) DEFAULT 'to-read',
    priority VARCHAR(20) CHECK (priority IN ('low', 'medium', 'high')) DEFAULT 'medium',
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    estimated_time INTEGER, -- in minutes
    progress INTEGER CHECK (progress >= 0 AND progress <= 100) DEFAULT 0,
    notes TEXT,
    tags TEXT[], -- PostgreSQL array for tags
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create learning_sessions table for tracking study sessions
CREATE TABLE IF NOT EXISTS learning_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_id UUID REFERENCES learning_resources(id) ON DELETE CASCADE,
    duration_minutes INTEGER NOT NULL,
    notes TEXT,
    session_date DATE DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create learning_goals table
CREATE TABLE IF NOT EXISTS learning_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    technology VARCHAR(100) NOT NULL,
    target_date DATE,
    status VARCHAR(20) CHECK (status IN ('active', 'completed', 'paused', 'cancelled')) DEFAULT 'active',
    progress INTEGER CHECK (progress >= 0 AND progress <= 100) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX idx_learning_resources_technology ON learning_resources(technology);
CREATE INDEX idx_learning_resources_type ON learning_resources(type);
CREATE INDEX idx_learning_resources_status ON learning_resources(status);
CREATE INDEX idx_learning_resources_priority ON learning_resources(priority);
CREATE INDEX idx_learning_resources_rating ON learning_resources(rating);
CREATE INDEX idx_learning_resources_created_at ON learning_resources(created_at);
CREATE INDEX idx_learning_resources_tags ON learning_resources USING GIN(tags);

CREATE INDEX idx_learning_sessions_resource_id ON learning_sessions(resource_id);
CREATE INDEX idx_learning_sessions_date ON learning_sessions(session_date);

CREATE INDEX idx_learning_goals_technology ON learning_goals(technology);
CREATE INDEX idx_learning_goals_status ON learning_goals(status);
CREATE INDEX idx_learning_goals_target_date ON learning_goals(target_date);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_learning_resources_updated_at 
    BEFORE UPDATE ON learning_resources 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_learning_goals_updated_at 
    BEFORE UPDATE ON learning_goals 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to automatically set completed_at when status changes to 'completed'
CREATE OR REPLACE FUNCTION set_completed_at()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' AND OLD.status != 'completed' THEN
        NEW.completed_at = NOW();
        NEW.progress = 100;
    ELSIF NEW.status != 'completed' THEN
        NEW.completed_at = NULL;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER set_learning_resource_completed_at
    BEFORE UPDATE ON learning_resources
    FOR EACH ROW
    EXECUTE FUNCTION set_completed_at();
