#!/bin/bash

# scripts/wait-for-services.sh

# Exit immediately if a command exits with a non-zero status.
set -e

# Kiểm tra xem file .env.test có tồn tại không và load nó
ENV_TEST_FILE="./.env.test" # Đường dẫn tương đối từ thư mục gốc của repo

if [ -f "$ENV_TEST_FILE" ]; then
  echo "Loading environment variables from $ENV_TEST_FILE"
  # Đọc file .env.test, loại bỏ comment đầu dòng, comment cuối dòng, và dòng trống
  # Sau đó export các biến KEY=VALUE
  while IFS= read -r line || [[ -n "$line" ]]; do
    # Loại bỏ comment cuối dòng (bất cứ thứ gì sau dấu #)
    cleaned_line=$(echo "$line" | sed 's/#.*//')
    # Loại bỏ khoảng trắng thừa ở đầu và cuối
    trimmed_line=$(echo "$cleaned_line" | sed 's/^[ \t]*//;s/[ \t]*$//')
    # Nếu dòng không rỗng và không phải là comment đầu dòng và chứa dấu =
    if [[ -n "$trimmed_line" && ! "$trimmed_line" =~ ^# && "$trimmed_line" == *"="* ]]; then
      export "$trimmed_line"
    fi
  done < "$ENV_TEST_FILE"
else
  echo "Error: $ENV_TEST_FILE not found."
  exit 1
fi

# --- Chờ MySQL ---
echo "Waiting for MySQL..."
max_attempts=${WAIT_MAX_ATTEMPTS:-12} # Sử dụng biến từ .env.test hoặc mặc định là 12
attempt_num=1

# Sử dụng MYSQL_PASSWORD và MYSQL_USER từ .env.test
# MYSQL_CONTAINER_NAME có thể lấy từ .env.test hoặc dùng giá trị mặc định
MYSQL_CONTAINER_NAME_FROM_ENV=${MYSQL_CONTAINER_NAME:-realtime_chat_app_mysql_test}

# MYSQL_USER_FOR_PING thường là 'root' cho mysqladmin ping, hoặc user có quyền tương đương
# Nếu .env.test có MYSQL_ROOT_USER thì dùng, không thì mặc định là root
MYSQL_PING_USER=${MYSQL_ROOT_USER:-root}
# MYSQL_PASSWORD_FOR_PING là mật khẩu của MYSQL_PING_USER
# Nếu .env.test có MYSQL_ROOT_PASSWORD thì dùng, không thì dùng MYSQL_PASSWORD
MYSQL_PING_PASSWORD=${MYSQL_ROOT_PASSWORD:-$MYSQL_PASSWORD}


until docker exec "$MYSQL_CONTAINER_NAME_FROM_ENV" mysqladmin ping -h localhost -u "$MYSQL_PING_USER" -p"${MYSQL_PING_PASSWORD}" || [ $attempt_num -eq $max_attempts ]; do
  >&2 echo "MySQL is unavailable (attempt $attempt_num/$max_attempts) on container '$MYSQL_CONTAINER_NAME_FROM_ENV' - sleeping 5s"
  sleep "${WAIT_SLEEP_INTERVAL:-5}" # Sử dụng biến từ .env.test hoặc mặc định là 5
  attempt_num=$((attempt_num+1))
done

if [ $attempt_num -eq $max_attempts ]; then
  >&2 echo "MySQL did not become healthy after $max_attempts attempts on container '$MYSQL_CONTAINER_NAME_FROM_ENV'."
  docker-compose -f docker-compose.test.yml logs "$MYSQL_CONTAINER_NAME_FROM_ENV"
  exit 1
fi
>&2 echo "MySQL on container '$MYSQL_CONTAINER_NAME_FROM_ENV' is up."

# --- Chờ Redis ---
echo "Waiting for Redis..."
attempt_num=1
# REDIS_CONTAINER_NAME có thể lấy từ .env.test hoặc dùng giá trị mặc định
REDIS_CONTAINER_NAME_FROM_ENV=${REDIS_CONTAINER_NAME:-realtime_chat_app_redis_test}
# Sử dụng REDIS_PASSWORD từ .env.test

until docker exec "$REDIS_CONTAINER_NAME_FROM_ENV" redis-cli ${REDIS_HOST:+-h "$REDIS_HOST"} ${REDIS_PORT:+-p "$REDIS_PORT"} -a "${REDIS_PASSWORD}" ping || [ $attempt_num -eq $max_attempts ]; do
  >&2 echo "Redis is unavailable (attempt $attempt_num/$max_attempts) on container '$REDIS_CONTAINER_NAME_FROM_ENV' - sleeping 5s"
  sleep "${WAIT_SLEEP_INTERVAL:-5}"
  attempt_num=$((attempt_num+1))
done

if [ $attempt_num -eq $max_attempts ]; then
  >&2 echo "Redis did not become healthy after $max_attempts attempts on container '$REDIS_CONTAINER_NAME_FROM_ENV'."
  docker-compose -f docker-compose.test.yml logs "$REDIS_CONTAINER_NAME_FROM_ENV"
  exit 1
fi
>&2 echo "Redis on container '$REDIS_CONTAINER_NAME_FROM_ENV' is up."

>&2 echo "All services are up - executing next steps."