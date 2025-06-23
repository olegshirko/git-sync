#!/bin/sh

GITLAB_URL="http://gitlab.local"
ROOT_PASSWORD_FILE="/var/opt/gitlab/initial_root_password"
ROOT_PASSWORD=""
ROOT_TOKEN=""
PROJECT_NAME="autoproject"
TOKEN_FILE="/scripts/generated-token.txt"

# Получаем root-пароль
echo "🔐 Извлечение root-пароля..."
while [ ! -f "$ROOT_PASSWORD_FILE" ]; do
  echo "⏳ Ожидание генерации пароля..."
  sleep 5
done

ROOT_PASSWORD=$(grep 'Password:' "$ROOT_PASSWORD_FILE" | awk '{print $2}')
echo "🔑 Пароль получен."

# Получаем root-токен
ROOT_TOKEN=$(curl -s --request POST "$GITLAB_URL/api/v4/session" \
  --form "login=root" \
  --form "password=$ROOT_PASSWORD" \
  | grep -o '"private_token":"[^"]*"' | cut -d':' -f2 | tr -d '"')

if [ -z "$ROOT_TOKEN" ]; then
  echo "❌ Не удалось получить root-токен."
  exit 1
fi
echo "✅ Root-токен получен."

# Создаём проект
echo "🚀 Создание проекта '$PROJECT_NAME'..."
curl -s --header "PRIVATE-TOKEN: $ROOT_TOKEN" \
  --data "name=$PROJECT_NAME" "$GITLAB_URL/api/v4/projects" > /dev/null

# Создаём токен доступа
echo "🎟️  Создание Personal Access Token..."
EXP_DATE=$(date -u -d "+30 days" +%Y-%m-%d)
TOKEN_JSON=$(curl -s --request POST "$GITLAB_URL/api/v4/personal_access_tokens" \
  --header "PRIVATE-TOKEN: $ROOT_TOKEN" \
  --form "name=automation-token" \
  --form "scopes[]=api" \
  --form "expires_at=$EXP_DATE")

ACCESS_TOKEN=$(echo "$TOKEN_JSON" | grep -o '"token":"[^"]*"' | cut -d':' -f2 | tr -d '"')

if [ -z "$ACCESS_TOKEN" ]; then
  echo "❌ Не удалось создать токен."
  exit 1
fi

# Сохраняем токен в файл
echo "$ACCESS_TOKEN" > "$TOKEN_FILE"
echo "✅ Токен сохранён в $TOKEN_FILE"
