#!/bin/sh

GITLAB_URL="http://gitlab.local"

echo "⏳ Жду, пока GitLab поднимется..."

until curl -s --head "$GITLAB_URL/users/sign_in" | grep "200 OK" > /dev/null; do
  sleep 5
done

echo "✅ GitLab готов!"
