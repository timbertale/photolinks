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
	<meta name="viewport" content="width=device-width, initial-scale=1.0">

	<title>Привязка ссылки</title>

	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
	<link href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:wght@400;500;600&display=swap" rel="stylesheet">

	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}

		body {
    		min-height: 100vh;

    		padding: 40px 20px;

    		font-family: "Cormorant Garamond", serif;

    		background:
        		linear-gradient(
            		rgba(255,255,255,.15),
            		rgba(255,255,255,.15)
        		),
        		url('/static/wood.jpg') center center / cover no-repeat;
		}

		.page {
    		min-height: 100vh;

    		display: flex;
    		flex-direction: column;
    		align-items: center;

    		padding: 40px 0;

		    gap: 24px;
		}

		.center {
    		flex: 1;

    		display: flex;
    		justify-content: center;
    		align-items: center;

    		width: 100%;
		}
		
		.container {
			width: min(90%, 850px);

			display: flex;
			flex-direction: column;
			align-items: center;

			gap: 30px;

			padding: 60px;

			background: rgba(255,255,255,.18);

			backdrop-filter: blur(8px);

			border-radius: 32px;

			border: 1px solid rgba(255,255,255,.35);

			box-shadow:
				0 25px 60px rgba(0,0,0,.15),
				inset 0 1px 0 rgba(255,255,255,.5);
		}

		h1 {
			text-align: center;

			font-size: clamp(2rem, 4vw, 4rem);

			font-weight: 500;

			line-height: 1.15;

			color: #5b3822;
		}

		form {
			width: 100%;

			display: flex;
			flex-direction: column;

			gap: 24px;
		}

		input {
			width: 100%;
			height: 78px;

			padding: 0 28px;

			border: none;
			outline: none;

			border-radius: 18px;

			font-size: 1.3rem;

			background: rgba(255,255,255,.85);

			box-shadow:
				inset 0 2px 8px rgba(0,0,0,.08),
				0 6px 16px rgba(0,0,0,.08);
		}

		input::placeholder {
			color: #9a816e;
		}

		button {
  		  width: 100%;
   		 height: 70px;

    border: none;
    border-radius: 18px;

    cursor: pointer;

    font-family: inherit;
    font-size: 1.6rem;

    color: white;

    background:
        linear-gradient(
            180deg,
            #c7965e,
            #9c6536
        );

    box-shadow:
        0 8px 20px rgba(120,75,35,.35),
        inset 0 1px 0 rgba(255,255,255,.4);

    transition: all .25s ease;
}

		button:hover {
			transform: translateY(-2px);

			box-shadow:
				0 12px 28px rgba(120,75,35,.4),
				inset 0 1px 0 rgba(255,255,255,.4);
		}

		button:active {
			transform: translateY(1px);
		}

		.logo {
    		width: clamp(120px, 20vw, 220px);
    		height: auto;

    		object-fit: contain;

    	filter:
        	drop-shadow(0 6px 14px rgba(0,0,0,.18));
		}

		@media (max-width: 768px) {

        body {
            padding: 20px;
        }

        .container {
            width: 100%;
            padding: 30px 20px;
            border-radius: 24px;
            gap: 24px;
        }

        h1 {
            font-size: 2rem;
            line-height: 1.2;
        }

        input {
            height: 60px;
            padding: 0 18px;
            font-size: 1.1rem;
        }

        button {
            width: 100%;
            min-width: auto;
            height: 60px;
            font-size: 1.3rem;
        }

		.page {
        	padding: 24px 0;
        	gap: 18px;
    	}

    	.center {
        	flex: none;
        	margin-top: auto;
        	margin-bottom: auto;
    	}
    }

    @media (max-width: 400px) {

        .container {
            padding: 24px 16px;
        }

        h1 {
            font-size: 1.7rem;
        }

        input,
        button {
            height: 54px;
            font-size: 1rem;
        }
    }
	</style>
</head>
<body>
<div class="page">

    <img src="/static/logo.png" alt="Logo" class="logo">

	<div class="center">
		<div class="container">
	
			<h1>
				Введите вашу ссылку<br>
				на облачное хранилище
			</h1>

			<form method="POST">
				<input
					type="url"
					name="storage_link"
					placeholder="https://..."
					required
				>

				<button type="submit">
					Сохранить
				</button>
			</form>

		</div>
	</div>
</div>
</body>
</html>
`))

func main() {
	http.Handle(
	"/static/",
	http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static")),
	),
)
	
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
