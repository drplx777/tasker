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

2. Авторизация пользователя
curl -X POST http://localhost:3000/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "ivanov",
    "password": "strongpassword"
  }'

3. Валидация токена
curl -X GET "http://localhost:3000/api/validate?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."


4. Получение информации о текущем пользователе
curl -X GET http://localhost:3000/api/getuserbyJWT \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."


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


2. Получение списка задач
curl -X GET http://localhost:3000/list \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."


3. Получение задачи по ID
curl -X GET http://localhost:3000/task/by_id/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

4. Обновление задачи
curl -X PUT http://localhost:3000/update/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Обновленный заголовок",
    "description": "Обновленное описание"
  }'


5. Удаление задачи
curl -X DELETE http://localhost:3000/delete/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

6. Отметка задачи как выполненной
curl -X PUT http://localhost:3000/done/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
Тестовые данные

Получение моковых задач
curl -X GET http://localhost:3000/tasklist