Аутентификация

1. Регистрация пользователя
curl -X POST http://localhost:3000/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Иван",
    "surname": "Иванов",
    "middlename": "Иванович",
    "login": "ivanov",
    "roleID": 1,
    "password": "strongpassword"
  }'

Responce

{
  "id": 123,
  "name": "Иван",
  "surname": "Иванов",
  "middlename": "Иванович",
  "login": "ivanov",
  "roleID": 1
}

2. Авторизация пользователя
curl -X POST http://localhost:3000/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "ivanov",
    "password": "strongpassword"
  }'

Responce
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 123,
    "name": "Иван",
    "surname": "Иванов",
    "login": "ivanov",
    "roleID": 1
  }
}

3. Валидация токена
curl -X GET "http://localhost:3000/api/validate?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

responce
{
  "valid": true,
  "userID": 123,
  "userRole": 1
}

4. Получение информации о текущем пользователе
curl -X GET http://localhost:3000/api/getuserbyJWT \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
Responce
{
  "id": 123,
  "name": "Иван",
  "surname": "Иванов",
  "login": "ivanov",
  "roleID": 1
}

Задачи (требуют аутентификации)
1. Создание задачи
curl -X POST http://localhost:3000/create \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Новая задача",
    "description": "Описание задачи",
    "status": "todo",
    "reporterId": "user-1",
    "deadline": "2025-12-31",
    "dashboardId": "dash-1"
  }'
responce
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Новая задача",
  "description": "Описание задачи",
  "status": "todo",
  "reporterId": "user-1",
  "createdAt": "2023-10-01T12:00:00Z",
  "updatedAt": "2023-10-01T12:00:00Z",
  "deadline": "2025-12-31",
  "dashboardId": "dash-1"
}

2. Получение списка задач
curl -X GET http://localhost:3000/list \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
responce
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Задача 1",
    "status": "in-progress",
    "createdAt": "2023-10-01T12:00:00Z"
  },
  {
    "id": "110e8400-e29b-41d4-a716-446655441111",
    "title": "Задача 2",
    "status": "done",
    "createdAt": "2023-10-02T12:00:00Z"
  }
]

3. Получение задачи по ID
curl -X GET http://localhost:3000/task/by_id/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
responce
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Детальная задача",
  "description": "Полное описание",
  "status": "in-progress",
  "reporterId": "user-1",
  "assignerId": "user-2",
  "createdAt": "2023-10-01T12:00:00Z",
  "updatedAt": "2023-10-01T14:00:00Z",
  "deadline": "2023-12-31",
  "dashboardId": "dash-1"
}


4. Обновление задачи
curl -X PUT http://localhost:3000/update/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Обновленный заголовок",
    "description": "Обновленное описание"
  }'
responce
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Обновленный заголовок",
  "description": "Обновленное описание",
  "status": "in-progress",
  "reporterId": "user-1",
  "assignerId": "user-2",
  "createdAt": "2023-10-01T12:00:00Z",
  "updatedAt": "2023-10-01T15:00:00Z",
  "deadline": "2023-12-31",
  "dashboardId": "dash-1"
}


5. Удаление задачи
curl -X DELETE http://localhost:3000/delete/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

6. Отметка задачи как выполненной
curl -X PUT http://localhost:3000/done/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
responce
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Выполненная задача",
  "description": "Описание",
  "status": "done",
  "reporterId": "user-1",
  "createdAt": "2023-10-01T12:00:00Z",
  "updatedAt": "2023-10-01T16:00:00Z",
  "completedAt": "2023-10-01T16:00:00Z",
  "deadline": "2023-12-31",
  "dashboardId": "dash-1"
}

Тестовые данные

Получение моковых задач
curl -X GET http://localhost:3000/tasklist