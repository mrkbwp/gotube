-- migrations/001_init_schema.sql

-- +goose Up
-- Создание расширения для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
                                     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                     username VARCHAR(50) NOT NULL,
                                     email VARCHAR(100) NOT NULL UNIQUE,
                                     password_hash VARCHAR(255) NOT NULL,
                                     avatar VARCHAR(255),
                                     role VARCHAR(20) NOT NULL DEFAULT 'user',
                                     created_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                     updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                     deleted_at TIMESTAMP WITH TIME ZONE,
                                     is_blocked BOOLEAN NOT NULL DEFAULT false,
                                     is_verified BOOLEAN NOT NULL DEFAULT false,
                                     verification_token VARCHAR(255),
                                     verification_expires_at TIMESTAMP WITH TIME ZONE,
                                     reset_token VARCHAR(255),
                                     reset_expires_at TIMESTAMP WITH TIME ZONE,
                                     metadata JSONB
);

-- Таблица для хранения refresh токенов
CREATE TABLE IF NOT EXISTS tokens (
                                      id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                      refresh_token TEXT NOT NULL UNIQUE,
                                      user_agent TEXT NOT NULL,
                                      client_ip TEXT NOT NULL,
                                      is_blocked BOOLEAN NOT NULL DEFAULT false,
                                      expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                      created_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                      deleted_at TIMESTAMP WITH TIME ZONE,
                                      metadata JSONB
);

-- Индексы для таблицы токенов
CREATE INDEX IF NOT EXISTS idx_tokens_refresh_token ON tokens(refresh_token);
CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_deleted_at ON tokens(deleted_at);

-- Таблица категорий
CREATE TABLE IF NOT EXISTS categories (
                                          id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                          name VARCHAR(50) NOT NULL UNIQUE,
                                          description TEXT,
                                          icon VARCHAR(255),
                                          created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                          updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                          deleted_at TIMESTAMP WITH TIME ZONE,
                                          is_active BOOLEAN NOT NULL DEFAULT true,
                                          sort_order INTEGER NOT NULL DEFAULT 0,
                                          parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
                                          metadata JSONB
);

-- Таблица видео
CREATE TABLE IF NOT EXISTS videos (
                                      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                      video_code VARCHAR(11) NOT NULL UNIQUE,
                                      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                      title VARCHAR(100) NOT NULL,
                                      description TEXT,
                                      category_id UUID REFERENCES categories(id) ON DELETE SET NULL,

    -- Пути хранения
                                      bucket_id VARCHAR(50) NOT NULL,
                                      shard_id VARCHAR(10) NOT NULL,
                                      path_segment1 VARCHAR(10) NOT NULL,
                                      path_segment2 VARCHAR(10) NOT NULL,
                                      filename VARCHAR(255) NOT NULL,

                                      thumbnail_url VARCHAR(255),
                                      duration INTEGER,
                                      views INTEGER NOT NULL DEFAULT 0,
                                      likes INTEGER NOT NULL DEFAULT 0,
                                      dislikes INTEGER NOT NULL DEFAULT 0,
                                      status VARCHAR(20) NOT NULL,

                                      is_blocked BOOLEAN NOT NULL DEFAULT false,
                                      is_private BOOLEAN NOT NULL DEFAULT false,
                                      processed_at TIMESTAMP WITH TIME ZONE,
                                      error_message TEXT,

                                      metadata JSONB,
                                      original_filename VARCHAR(255),

                                      created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                      updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                      deleted_at TIMESTAMP WITH TIME ZONE,

                                      CHECK (char_length(video_code) = 11)
);

CREATE INDEX IF NOT EXISTS idx_videos_user_id ON videos(user_id);
CREATE INDEX IF NOT EXISTS idx_videos_category_id ON videos(category_id);
CREATE INDEX IF NOT EXISTS idx_videos_created_at ON videos(created_at);
CREATE INDEX IF NOT EXISTS idx_videos_views ON videos(views);
CREATE INDEX IF NOT EXISTS idx_videos_deleted_at ON videos(deleted_at);
CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
CREATE INDEX IF NOT EXISTS idx_videos_is_blocked ON videos(is_blocked);
CREATE INDEX IF NOT EXISTS idx_videos_is_private ON videos(is_private);
CREATE UNIQUE INDEX IF NOT EXISTS udx__videos__video__code ON videos(video_code);

-- Таблица реакций на видео (лайки/дизлайки)
CREATE TABLE IF NOT EXISTS video_reactions (
                                               id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                               video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
                                               user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                               type VARCHAR(10) NOT NULL, -- 'like' или 'dislike'
                                               created_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                               deleted_at TIMESTAMP WITH TIME ZONE,
                                               UNIQUE(video_id, user_id)
);

-- Индексы для таблицы реакций
CREATE INDEX IF NOT EXISTS idx_video_reactions_video_id ON video_reactions(video_id);
CREATE INDEX IF NOT EXISTS idx_video_reactions_user_id ON video_reactions(user_id);
CREATE INDEX IF NOT EXISTS idx_video_reactions_deleted_at ON video_reactions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_video_reactions_type ON video_reactions(type);

-- Таблица доступных качеств видео
CREATE TABLE IF NOT EXISTS video_qualities (
                                               id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                               name VARCHAR(20) NOT NULL UNIQUE,
                                               width INTEGER NOT NULL,
                                               height INTEGER NOT NULL,
                                               target_bitrate INTEGER NOT NULL,
                                               created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                               is_active BOOLEAN NOT NULL DEFAULT true
);

-- Таблица файлов видео разных качеств и форматов
CREATE TABLE IF NOT EXISTS video_files (
                                           id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                           video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
                                           quality_id UUID NOT NULL REFERENCES video_qualities(id),

                                           file_format VARCHAR(10) NOT NULL, -- mp4, webm и т.д.
                                           file_size BIGINT,

                                           width INTEGER NOT NULL,
                                           height INTEGER NOT NULL,
                                           bitrate INTEGER,

                                           status VARCHAR(20) NOT NULL DEFAULT 'pending',
                                           created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                           updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

                                           UNIQUE(video_id, quality_id, file_format)
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_video_files_video_id ON video_files(video_id);
CREATE INDEX IF NOT EXISTS idx_video_files_quality_id ON video_files(quality_id);
CREATE INDEX IF NOT EXISTS idx_video_files_status ON video_files(status);

-- Базовые качества
INSERT INTO video_qualities (name, width, height, target_bitrate) VALUES
                                                                      ('240p', 352, 240, 400000),
                                                                      ('360p', 480, 360, 800000),
                                                                      ('480p', 854, 480, 1500000),
                                                                      ('720p', 1280, 720, 2500000),
                                                                      ('1080p', 1920, 1080, 4500000),
                                                                      ('4K', 3840, 2160, 15000000)
ON CONFLICT (name) DO NOTHING;

-- Таблица комментариев
CREATE TABLE IF NOT EXISTS comments (
                                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                        video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
                                        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                        parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
                                        text TEXT NOT NULL,
                                        created_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                        updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                        deleted_at TIMESTAMP WITH TIME ZONE,
                                        is_blocked BOOLEAN NOT NULL DEFAULT false,
                                        likes INTEGER NOT NULL DEFAULT 0,
                                        dislikes INTEGER NOT NULL DEFAULT 0,
                                        metadata JSONB
);

-- Индексы для таблицы комментариев
CREATE INDEX IF NOT EXISTS idx_comments_video_id ON comments(video_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id);
CREATE INDEX IF NOT EXISTS idx_comments_deleted_at ON comments(deleted_at);
CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at);
CREATE INDEX IF NOT EXISTS idx_comments_is_blocked ON comments(is_blocked);

-- Таблица реакций на комментарии
CREATE TABLE IF NOT EXISTS comment_reactions (
                                                 id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                                 comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
                                                 user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                                 type VARCHAR(10) NOT NULL, -- 'like' или 'dislike'
                                                 created_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                                 deleted_at TIMESTAMP WITH TIME ZONE,
                                                 UNIQUE(comment_id, user_id)
);

-- Индексы для таблицы реакций на комментарии
CREATE INDEX IF NOT EXISTS idx_comment_reactions_comment_id ON comment_reactions(comment_id);
CREATE INDEX IF NOT EXISTS idx_comment_reactions_user_id ON comment_reactions(user_id);
CREATE INDEX IF NOT EXISTS idx_comment_reactions_deleted_at ON comment_reactions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_comment_reactions_type ON comment_reactions(type);

-- Таблица подписок
CREATE TABLE IF NOT EXISTS subscriptions (
                                             id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                             subscriber_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                             channel_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                             created_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                             deleted_at TIMESTAMP WITH TIME ZONE,
                                             UNIQUE(subscriber_id, channel_id)
);

-- Индексы для таблицы подписок
CREATE INDEX IF NOT EXISTS idx_subscriptions_subscriber_id ON subscriptions(subscriber_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_channel_id ON subscriptions(channel_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_deleted_at ON subscriptions(deleted_at);

-- Таблица плейлистов
CREATE TABLE IF NOT EXISTS playlists (
                                         id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                         user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                         title VARCHAR(100) NOT NULL,
                                         description TEXT,
                                         is_private BOOLEAN NOT NULL DEFAULT false,
                                         created_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                         updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                         deleted_at TIMESTAMP WITH TIME ZONE,
                                         metadata JSONB
);

-- Индексы для таблицы плейлистов
CREATE INDEX IF NOT EXISTS idx_playlists_user_id ON playlists(user_id);
CREATE INDEX IF NOT EXISTS idx_playlists_deleted_at ON playlists(deleted_at);
CREATE INDEX IF NOT EXISTS idx_playlists_is_private ON playlists(is_private);

-- Таблица видео в плейлистах
CREATE TABLE IF NOT EXISTS playlist_videos (
                                               id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                               playlist_id UUID NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
                                               video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
                                               position INTEGER NOT NULL,
                                               created_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                               deleted_at TIMESTAMP WITH TIME ZONE,
                                               UNIQUE(playlist_id, video_id)
);

-- Индексы для таблицы видео в плейлистах
CREATE INDEX IF NOT EXISTS idx_playlist_videos_playlist_id ON playlist_videos(playlist_id);
CREATE INDEX IF NOT EXISTS idx_playlist_videos_video_id ON playlist_videos(video_id);
CREATE INDEX IF NOT EXISTS idx_playlist_videos_deleted_at ON playlist_videos(deleted_at);
CREATE INDEX IF NOT EXISTS idx_playlist_videos_position ON playlist_videos(position);

-- Заполнение начальных данных для категорий
INSERT INTO categories (
    id,
    name,
    description,
    icon,
    created_at,
    updated_at,
    is_active,
    sort_order
) VALUES
      (uuid_generate_v4(), 'Развлечения', 'Развлекательные видео', 'entertainment.svg', NOW(), NOW(), true, 1),
      (uuid_generate_v4(), 'Обучение', 'Обучающие видео', 'education.svg', NOW(), NOW(), true, 2),
      (uuid_generate_v4(), 'Игры', 'Игровые видео', 'gaming.svg', NOW(), NOW(), true, 3),
      (uuid_generate_v4(), 'Музыка', 'Музыкальные видео', 'music.svg', NOW(), NOW(), true, 4),
      (uuid_generate_v4(), 'Наука и технологии', 'Научные и технологические видео', 'science.svg', NOW(), NOW(), true, 5)
ON CONFLICT (name) DO NOTHING;
