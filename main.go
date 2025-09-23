package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

// HTML-шаблон формы
var formTmpl = template.Must(template.New("form").Parse(`
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Привязка ссылки</title>
</head>
<body>
	<h1>Введите вашу ссылку на облако</h1>
	<form method="POST">
		<input type="url" name="storage_link" placeholder="https://..." required style="width:300px">
		<button type="submit">Сохранить</button>
	</form>
</body>
</html>
`))

func main() {
	var err error

	// Получаем строку подключения к базе (Render задаёт DATABASE_URL в переменных окружения)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL не установлена")
	}

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаём таблицу, если ещё нет
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS storage_links (
		uid TEXT PRIMARY KEY,
		link TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	)
	`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Сервер запущен на :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		http.Error(w, "uid обязателен", http.StatusBadRequest)
		return
	}

	// Проверяем, есть ли запись
	var link string
	err := db.QueryRow("SELECT link FROM storage_links WHERE uid = $1", uid).Scan(&link)

	if r.Method == http.MethodGet {
		if err == sql.ErrNoRows {
			// Нет записи → показываем форму
			formTmpl.Execute(w, nil)
			return
		} else if err != nil {
			http.Error(w, "Ошибка БД", http.StatusInternalServerError)
			return
		}

		// Есть запись → редирект
		http.Redirect(w, r, link, http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		if err == sql.ErrNoRows {
			// Читаем ссылку
			storageLink := r.FormValue("storage_link")
			if storageLink == "" {
				http.Error(w, "ссылка обязательна", http.StatusBadRequest)
				return
			}

			// Сохраняем
			_, err := db.Exec("INSERT INTO storage_links (uid, link) VALUES ($1, $2)", uid, storageLink)
			if err != nil {
				http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, storageLink, http.StatusFound)
			return
		}

		// Если запись уже существует → редиректим
		http.Redirect(w, r, link, http.StatusFound)
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}
