![License](https://img.shields.io/badge/license-BUSL--1.1-blue.svg)

**GoTube - Видеохостинг**

GoTube - это современная платформа для хостинга видео, разработанная на Go. Предоставляет функционал, схожий с YouTube, позволяя пользователям загружать, делиться и просматривать видео.

**Демо**: https://gotube.cc

**Основные возможности**
- Загрузка и обработка видео (ffmpeg)
- Множество вариантов качества (240p, 360p, 480p, 720p, 1080p, 4k), также можно добавить дополнительные
- Вход и регистрация пользователей
- Система комментариев
- Лайки и дизлайки
- Расчет на масштабирование
- Категории видео

**Технологии**
- **Бэкенд**: Go (Echo Framework)
- **База данных**: PostgreSQL
- **Кэширование**: Redis
- **Хранилище**: MinIO
- **Очередь сообщений**: Kafka
- **Документация**: Swagger/OpenAPI
- **Контейнеризация**: Docker

**Требования**
- Go 1.23+
- Docker и Docker Compose
- PostgreSQL 15+
- Redis 7+
- MinIO
- Kafka

**Roadmap**
- Вынести обработку видео в отдельный сервис (сообщения в кафку отправляются)
- Вынести базу справочников (категории, качества видео) в отдельную базу и сервис
- Доработать построитель запросов
- Приватность видео
- Рекомендации видео
- Обработка видео ML
- Плейлисты и подписки на пользователей
- Оповещения
- 
**Быстрый старт**
- Клонируем репозиторий:
```bash
git clone github.com/mrkbwp/gotube/gotube.git
cd gotube
```
- Копируем файл с переменными окружения:
```
cp .env.example .env
```
- Запускаем сервисы через Docker Compose:
```
docker-compose up -d
```

- Запускаем приложение:
```
go run cmd/api/main.go
```

***Документация API***
Документация API доступна через Swagger UI по адресу:
```
http://localhost:8111/swagger/
```

***Генерация документации Swagger***
```
swag init -g cmd/api/main.go
```

***Участие в разработке***

- Форкните репозиторий
- Создайте ветку для новой функции (git checkout -b feature/amazing-feature)
- Зафиксируйте изменения (git commit -m 'Добавлена новая функция')
- Отправьте изменения в репозиторий (git push origin feature/amazing-feature)
- Создайте Pull Request

## Лицензия

Этот проект лицензирован под BUSL-1.1 - см. файлы [LICENSE](LICENSE) и [COMMERCIAL.md](COMMERCIAL.md) для подробностей.

- Разрешено: Просмотр кода, форки, использование для обучения
- Запрещено: Коммерческое использование без разрешения автора
- Для коммерческого использования: [свяжитесь с автором](COMMERCIAL.md)

***Авторы***
- Mrkbwp - Начальная разработка - https://github.com/mrkbwp/gotube
