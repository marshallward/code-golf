package main

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error

	if db, err = sql.Open("postgres", os.Getenv("DB_DSN")); err != nil {
		panic(err)
	}
}

func addSolution(userID int, lang, code string) {
	if _, err := db.Exec(
		`INSERT INTO solutions(user_id, lang, code) VALUES($1, $2, $3)
		 ON CONFLICT ON CONSTRAINT solutions_pkey DO UPDATE SET code = $3`,
		userID, lang, code,
	); err != nil {
		panic(err)
	}
}

func getSolutionCode(userID int, lang string) (code string) {
	db.QueryRow(
		"SELECT code FROM solutions WHERE user_id = $1 AND lang = $2",
		userID, lang,
	).Scan(&code)

	return
}

func addUser(id int, login string) {
	if _, err := db.Exec(
		`INSERT INTO users VALUES($1, $2)
		 ON CONFLICT(id) DO UPDATE SET login = $2`,
		id, login,
	); err != nil {
		panic(err)
	}
}

func ordinal(x int) string {
	suffix := "th"
	switch x % 10 {
	case 1:
		if x%100 != 11 {
			suffix = "st"
		}
	case 2:
		if x%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if x%100 != 13 {
			suffix = "rd"
		}
	}
	return strconv.Itoa(x) + "<sup>" + suffix + "</sup>"
}

func printLeaderboards(w http.ResponseWriter) {
	rows, err := db.Query(
		`SELECT login, lang, LENGTH(code)
		   FROM solutions
		   JOIN users on user_id = id
		  ORDER BY LENGTH(code)`,
	)

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	w.Write([]byte("<article><table><tr><th>Rank<th>Strokes<th>Lang<th>User"))

	i := 0

	for rows.Next() {
		var login, lang string
		var length int

		if err := rows.Scan(&login, &lang, &length); err != nil {
			panic(err)
		}

		i++

		w.Write([]byte(
			"<tr><td>" + ordinal(i) +
			"<td>" + strconv.Itoa(length) +
			"<td>" + strings.Title(lang) +
			`<td><img src="//avatars.githubusercontent.com/` + login +
			`?size=20"> ` + login,
		))
	}

	w.Write([]byte("</table></article>"))

	if err := rows.Err(); err != nil {
		panic(err)
	}
}
