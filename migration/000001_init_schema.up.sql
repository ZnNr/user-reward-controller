-- Удаление таблиц в обратном порядке
DROP TABLE IF EXISTS UserVisits CASCADE;
DROP TABLE IF EXISTS UserActivityLog CASCADE;
DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS task_status CASCADE;
DROP TABLE IF EXISTS referral CASCADE;
DROP TABLE IF EXISTS Users CASCADE;

-- Создание таблицы пользователей
CREATE TABLE Users (
                       ID VARCHAR(255) PRIMARY KEY NOT NULL,
                       Username VARCHAR(255) NOT NULL,
                       Email VARCHAR(255) NOT NULL UNIQUE,
                       Balance DECIMAL(15, 2) CHECK (Balance >= 0),
                       Referrals INT DEFAULT 0 CHECK (Referrals >= 0),
                       ReferralCode VARCHAR(50),
                       TasksCompleted INT DEFAULT 0 CHECK (TasksCompleted >= 0),
                       CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
                       UpdatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
                       LastVisit TIMESTAMP,
                       VisitCount INT DEFAULT 0,
                       Bio TEXT,
                       TimeZone VARCHAR(50),
                       Status VARCHAR(20) NOT NULL CHECK (Status IN ('active', 'inactive', 'banned'))
);

-- Создание таблицы для логов активности пользователей
CREATE TABLE UserActivityLog (
                                 ID SERIAL PRIMARY KEY,
                                 UserID VARCHAR(255) NOT NULL,
                                 ActivityTime TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
                                 FOREIGN KEY (UserID) REFERENCES Users(ID) ON DELETE CASCADE
);

-- Создание таблицы для посещений пользователей
CREATE TABLE UserVisits (
                            UserID VARCHAR(255) NOT NULL,
                            VisitDate TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
                            FOREIGN KEY (UserID) REFERENCES Users(ID) ON DELETE CASCADE,
                            PRIMARY KEY (UserID, VisitDate)
);

-- Создание таблицы статусов задач
CREATE TABLE task_status (
                             id SERIAL PRIMARY KEY,
                             name VARCHAR(50) NOT NULL UNIQUE
);

-- Вставка начальных значений в таблицу статусов задач
INSERT INTO task_status (name) VALUES
                                   ('Not Started'),
                                   ('In Progress'),
                                   ('Completed'),
                                   ('Canceled');

-- Создание таблицы задач
CREATE TABLE tasks (
                       task_id VARCHAR(255) PRIMARY KEY NOT NULL,
                       title VARCHAR(255) NOT NULL,
                       description TEXT,
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       due_date TIMESTAMP WITH TIME ZONE,
                       status INT NOT NULL REFERENCES task_status(id),
                       assignee_id UUID
);

-- Создание таблицы рефералов
CREATE TABLE referral (
                          referral_id SERIAL PRIMARY KEY,
                          user_id VARCHAR(255) NOT NULL,
                          code VARCHAR(255) NOT NULL UNIQUE,
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для улучшения производительности при фильтрации
CREATE INDEX idx_tasks_title ON tasks(title);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_assignee ON tasks(assignee_id);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);