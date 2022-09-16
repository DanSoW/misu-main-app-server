package main

import (
	"context"
	"fmt"
	mainserver "main-server"
	config "main-server/config"
	handler "main-server/pkg/handler"
	repository "main-server/pkg/repository"
	service "main-server/pkg/service"
	"os"
	"os/signal"
	"syscall"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/spf13/viper"
)

// @title MISU Main Server
// @version 1.0
// description Аналитическое ядро системы МИСУ

// @host localhost:5000
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	// Инициализация конфигурации сервера
	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variable: %s", err.Error())
	}

	// Инициализация логгера
	logrus.SetFormatter(new(logrus.JSONFormatter))

	// Открытие файла для записи логгов
	fileError, err := os.OpenFile(viper.GetString("paths.logs.error"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		// Добавление хука для записи всех сообщений уровня logrus.ErrorLevel в открытый файл (запись ошибок в файл)
		logrus.AddHook(&writer.Hook{
			Writer: fileError,
			LogLevels: []logrus.Level{
				logrus.ErrorLevel,
			},
		})
	} else {
		// При ошибки открытия файла запись осуществляется в стандартную консоль
		logrus.SetOutput(os.Stderr)
		logrus.Error("Failed to log to file, using default stderr")
	}

	fileInfo, err := os.OpenFile(viper.GetString("paths.logs.info"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logrus.AddHook(&writer.Hook{
			Writer: fileInfo,
			LogLevels: []logrus.Level{
				logrus.InfoLevel,
				logrus.DebugLevel,
			},
		})
	} else {
		logrus.SetOutput(os.Stderr)
		logrus.Error("Failed to log to file, using default stderr")
	}

	fileWarn, err := os.OpenFile(viper.GetString("paths.logs.warn"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logrus.AddHook(&writer.Hook{
			Writer: fileWarn,
			LogLevels: []logrus.Level{
				logrus.WarnLevel,
			},
		})
	} else {
		logrus.SetOutput(os.Stderr)
		logrus.Error("Failed to log to file, using default stderr")
	}

	fileFatal, err := os.OpenFile(viper.GetString("paths.logs.fatal"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logrus.AddHook(&writer.Hook{
			Writer: fileFatal,
			LogLevels: []logrus.Level{
				logrus.FatalLevel,
			},
		})
	} else {
		logrus.SetOutput(os.Stderr)
		logrus.Error("Failed to log to file, using default stderr")
	}

	// Подключение к базе данных (основное подключение)
	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: os.Getenv("DB_PASSWORD"),
	})

	// Строка DNS, используемая при подключении к базе данных
	dns := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		viper.GetString("db.host"),
		viper.GetString("db.username"),
		os.Getenv("DB_PASSWORD"),
		viper.GetString("db.dbname"),
		viper.GetString("db.port"),
		viper.GetString("db.sslmode"),
	)

	// Подключение к базе данных с помощью ORM Gorm
	dbAdapter, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dns,
	}), &gorm.Config{})

	// Создание нового адаптера с помощью gormadapter, для получения информации о всех правилах содержащихся в таблице
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(dbAdapter, &config.MisuRule{}, viper.GetString("rules_table_name"))

	if err != nil {
		logrus.Fatalf("failed to initialize adapter by db with custom table: %s", err.Error())
	}

	// Получение информации о всех правилах содержащихся в БД (синхронизация PERM модели и существующих данных в БД)
	enforcer, err := casbin.NewEnforcer(viper.GetString("paths.perm_model"), adapter)

	if err != nil {
		logrus.Fatalf("failed to initialize new enforcer: %s", err.Error())
	}

	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}

	// Инициализация данных для доступа к внешним сервисам авторизации / регистрации
	config.InitOAuth2Config()

	// Реализация подхода dependency injection
	repos := repository.NewRepository(db, enforcer)
	service := service.NewService(repos)
	handlers := handler.NewHandler(service)

	// Создание нового экзепляра сервиса
	srv := new(mainserver.Server)

	// Запуск горутины, в котором запускается сервер
	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
			logrus.Fatalf("error occured while running http server: %s", err.Error())
		}
	}()

	logrus.Print("MISU Main Server Started")

	// Реализация Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("MISU Main Server Shutting Down")

	// Освобождение ресурсов сервера
	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		logrus.Errorf("error occured on db connection close: %s", err.Error())
	}

	// Закрытие файлов логгера
	if err := fileError.Close(); err != nil {
		logrus.Error(err.Error())
	}

	if err := fileWarn.Close(); err != nil {
		logrus.Error(err.Error())
	}

	if err := fileInfo.Close(); err != nil {
		logrus.Error(err.Error())
	}

	if err := fileFatal.Close(); err != nil {
		logrus.Error(err.Error())
	}
}

// Инициализация конфига
func initConfig() error {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}
