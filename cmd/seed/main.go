package main

import (
	"database/sql"
	"fmt"
	"gochat-backend/config"
	"gochat-backend/internal/infra/mysqlinfra"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/bxcodec/faker/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID        string
	Name      string
	AvatarURL string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var fakerMutex sync.Mutex

// Mảng chứa các URL avatar cố định
var avatarURLs = []string{
	"https://res.cloudinary.com/durc9hj8m/image/upload/v1745339952/titan_10_hvq29s.jpg",
	"https://res.cloudinary.com/durc9hj8m/image/upload/v1745339951/titan_7_hqp1sc.jpg",
	"https://res.cloudinary.com/durc9hj8m/image/upload/v1745339951/titan_9_yhlqyj.jpg",
	"https://res.cloudinary.com/durc9hj8m/image/upload/v1745339951/tiga_2_twvfjn.jpg",
	"https://res.cloudinary.com/durc9hj8m/image/upload/v1745339951/titan_5_ffyxpm.jpg",
	"https://res.cloudinary.com/durc9hj8m/image/upload/v1745339951/titan_8_v4oi7c.jpg",
	"https://res.cloudinary.com/durc9hj8m/image/upload/v1745339951/titan_6_ilrlka.jpg",
	"https://res.cloudinary.com/durc9hj8m/image/upload/v1745339951/tiga_3_yb2vul.jpg",
	"https://res.cloudinary.com/durc9hj8m/image/upload/v1745339950/tiga_1_sa6msh.jpg",
}

func generateFakeUser() Account {
	now := time.Now()
	// Lock the mutex before using faker
	fakerMutex.Lock()
	password := faker.Password()
	id := faker.UUIDDigit()
	name := faker.Name()
	// Chọn URL avatar ngẫu nhiên từ mảng
	avatarURL := avatarURLs[rand.Intn(len(avatarURLs))]
	email := faker.Email()
	fakerMutex.Unlock()

	// Hash password after unlocking the mutex to improve concurrency
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		// Fallback to plain password in case of error
		hashedPassword = []byte(password)
	}

	return Account{
		ID:        id,
		Name:      name,
		AvatarURL: avatarURL,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func insertBatch(db *sql.DB, accounts []Account) error {
	if len(accounts) == 0 {
		return nil
	}

	var builder strings.Builder
	builder.WriteString("INSERT INTO users (id, name, avatar_url, email, password, created_at, updated_at) VALUES ")

	args := make([]interface{}, 0, len(accounts)*7)

	for i, acc := range accounts {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("(?, ?, ?, ?, ?, ?, ?)")
		args = append(args,
			acc.ID, acc.Name, acc.AvatarURL, acc.Email, acc.Password, acc.CreatedAt, acc.UpdatedAt,
		)
	}

	stmt := builder.String()
	_, err := db.Exec(stmt, args...)
	return err
}

func main() {
	// Khởi tạo seed cho random để các avatar có thể xuất hiện đều nhau
	rand.Seed(time.Now().UnixNano())

	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := mysqlinfra.ConnectMysql(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	database := mysqlinfra.NewMySqlDatabase(db)
	defer database.Close()

	totalUsers := 50000
	batchSize := 1000
	numWorkers := 8 // Số lượng goroutine cần tạo

	start := time.Now()

	// Channel để nhận kết quả từ goroutines
	accountsChan := make(chan Account)
	// Channel để theo dõi khi nào tất cả goroutines đã hoàn thành
	done := make(chan bool)

	// Mutex để đảm bảo thread-safety khi chèn dữ liệu vào database
	var mutex sync.Mutex
	var insertedCount int
	var currentBatch []Account

	// Goroutine để thu thập và chèn dữ liệu
	go func() {
		for acc := range accountsChan {
			mutex.Lock()
			currentBatch = append(currentBatch, acc)

			if len(currentBatch) >= batchSize {
				batchToInsert := make([]Account, len(currentBatch))
				copy(batchToInsert, currentBatch)
				currentBatch = currentBatch[:0]
				insertedCount += len(batchToInsert)

				// Unblock mutex trước khi thực hiện insert
				mutex.Unlock()

				if err := insertBatch(db, batchToInsert); err != nil {
					log.Fatalf("Batch insert failed: %v", err)
				}

				fmt.Printf("Inserted %d users...\n", insertedCount)
			} else {
				mutex.Unlock()
			}
		}

		// Xử lý batch cuối cùng nếu còn
		mutex.Lock()
		if len(currentBatch) > 0 {
			if err := insertBatch(db, currentBatch); err != nil {
				log.Fatalf("Final batch insert failed: %v", err)
			}
			insertedCount += len(currentBatch)
			fmt.Printf("Inserted final batch. Total: %d users\n", insertedCount)
		}
		mutex.Unlock()

		done <- true
	}()

	// Tạo worker pool để generate fake users
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	usersPerWorker := totalUsers / numWorkers

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()

			start := workerID * usersPerWorker
			end := start + usersPerWorker
			if workerID == numWorkers-1 {
				// Đảm bảo worker cuối cùng xử lý hết số lượng còn lại
				end = totalUsers
			}

			for j := start; j < end; j++ {
				accountsChan <- generateFakeUser()
			}
		}(i)
	}

	// Đợi tất cả workers hoàn thành
	go func() {
		wg.Wait()
		close(accountsChan)
	}()

	// Đợi cho đến khi tất cả các batch đã được insert
	<-done

	fmt.Println("✅ Done! Total time:", time.Since(start))
}
