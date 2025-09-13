package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "modernc.org/sqlite" // драйвер SQLite без CGO
)

var db *sql.DB

// HTML-шаблон для формы
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

	// Подключаем базу (файл создастся сам)
	db, err = sql.Open("sqlite", "storage_links.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем таблицу, если еще нет
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS storage_links (
		uid TEXT PRIMARY KEY,
		link TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	log.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		http.Error(w, "uid обязателен", http.StatusBadRequest)
		return
	}

	// Проверяем в БД
	var link string
	err := db.QueryRow("SELECT link FROM storage_links WHERE uid = ?", uid).Scan(&link)

	if r.Method == http.MethodGet {
		if err == sql.ErrNoRows {
			// Записи нет → показываем форму
			formTmpl.Execute(w, nil)
			return
		} else if err != nil {
			http.Error(w, "Ошибка БД", http.StatusInternalServerError)
			return
		}

		// Запись найдена → редирект
		http.Redirect(w, r, link, http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		if err == sql.ErrNoRows {
			// Читаем ссылку из формы
			storageLink := r.FormValue("storage_link")
			if storageLink == "" {
				http.Error(w, "ссылка обязательна", http.StatusBadRequest)
				return
			}

			// Сохраняем в БД
			_, err := db.Exec("INSERT INTO storage_links (uid, link) VALUES (?, ?)", uid, storageLink)
			if err != nil {
				http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
				return
			}

			// Редиректим
			http.Redirect(w, r, storageLink, http.StatusFound)
			return
		}

		// Если запись уже существует → сразу редирект
		http.Redirect(w, r, link, http.StatusFound)
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}
