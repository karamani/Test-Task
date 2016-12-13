// Веб-служба, которая загружает прайс-лист с товарами в базу даных.
// Формат запроса для загрузки прайс-листа в БД:
// POST http://{domian.com}/price/{placeholderID}/upload
// В теле запроса должен содержаться файл формата csv,
// и содержать 3 колонки, разделенные табуляциями: art, count и price.
// Значения должны быть целочисленными.
// Формат запроса для получения товаров определенного прайс-листа:
// GET http://{domian.com}/price/{placeholderID}
// Параметры GET-запроса:
// skip - необязательный параметр - сколько товаров пропускать в выдаче
// limit - необязательный параметр - количество товаров на странице

package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	logErrorTCV = "Invalid .scv file"
	logErrorURL = "Invalid URL request"
)

type (
	Product struct {
		Art   int
		Count int
		Price int
	}
)

var (
	queryInsert *sql.Stmt
	querySelect *sql.Stmt
	queryDelete *sql.Stmt
)

func httpHandler(w http.ResponseWriter, r *http.Request) {
	//Проверки URL
	url := strings.Split(r.URL.Path, "/")
	if len(url) < 3 {
		log.Println(logErrorURL)
		return
	}
	distributorId, err := strconv.Atoi(url[2])
	if err != nil {
		log.Println(logErrorURL)
		return
	}
	if (len(url) < 4 || url[3] == "") && r.Method == http.MethodGet {
		//GET запрос
		//Проверка необязательных параметров запроса
		skip := 0
		paramSkip := r.URL.Query().Get("skip")
		if len(paramSkip) > 0 {
			skip, err = strconv.Atoi(paramSkip)
			if err != nil {
				skip = 0
			}
		}
		limit := -1
		paramLimit := r.URL.Query().Get("limit")
		if len(paramLimit) > 0 {
			limit, err = strconv.Atoi(paramLimit)
			if err != nil {
				limit = -1
			}
		}
		//Выбор строк из БД
		rows, err := querySelect.Query(distributorId, skip, limit)
		if err != nil {
			log.Println(err)
			return
		}
		defer rows.Close()
		//Чтение строк результата
		var prodList []Product
		var prod Product
		for rows.Next() {
			err = rows.Scan(&prod.Art, &prod.Count, &prod.Price)
			if err != nil {
				log.Println(err)
				return
			}
			prodList = append(prodList, prod)
		}
		//Перевод в JSON
		response, err := json.Marshal(prodList)
		if err != nil {
			log.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	} else if len(url) >= 4 && url[3] == "upload" && len(url) <= 5 && r.Method == http.MethodPost {
		//POST запрос
		//Получение файла из запроса
		file, _, err := r.FormFile("scv")
		if err != nil {
			log.Println(err)
			return
		}
		defer file.Close()
		//Удаление устаревших данных из таблицы (с указанным в URL distributorId)
		_, err = queryDelete.Exec(distributorId)
		if err != nil {
			log.Println(err)
			return
		}
		//Чтение файла из запроса
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			fields := strings.Split(scanner.Text(), "\t")
			//Проверки на корректность данных
			if len(fields) != 3 {
				log.Println(logErrorTCV)
				return
			}
			art, err := strconv.Atoi(fields[0])
			if err != nil {
				log.Println(logErrorTCV)
				return
			}
			count, err := strconv.Atoi(fields[1])
			if err != nil {
				log.Println(logErrorTCV)
				return
			}
			price, err := strconv.Atoi(fields[2])
			if err != nil {
				log.Println(logErrorTCV)
				return
			}
			//Запрос на добавление в БД
			_, err = queryInsert.Exec(distributorId, art, count, price)
			if err != nil {
				log.Println(err)
			}
		}
	} else {
		log.Println(logErrorURL)
	}
}

func main() {
	//Строка с параметрами соединения передается в параметрах командной строки
	connString := flag.String("conn", "sql7149110:tjlEzXXbvG@tcp(sql7.freemysqlhosting.net:3306)/sql7149110", "user:password@protocol(ip:port)/database")
	//Подключение к БД
	db, err := sql.Open("mysql", *connString)
	if err != nil {
		log.Fatal(err)
	}
	//Подготовка запросов
	queryDelete, err = db.Prepare("DELETE FROM pricelist WHERE distributorId = ?")
	if err != nil {
		log.Fatal(err)
	}
	queryInsert, err = db.Prepare("INSERT INTO pricelist VALUES (?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	querySelect, err = db.Prepare("SELECT art, count, price FROM pricelist WHERE distributorId = ? ORDER BY art LIMIT ?, ?")
	if err != nil {
		log.Fatal(err)
	}
	//Запуск HTTP сервера
	http.HandleFunc("/price/", httpHandler)
	http.ListenAndServe(":80", nil)
}
