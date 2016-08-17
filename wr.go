package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"io"
	"encoding/csv"
)

// main はエントリポイントです。
func main() {
	// コマンドラインから対象のディレクトリを取得
	dir := flag.String("d", "", "source directory path")
	pattern := flag.String("p", "", "match pattern")
	setting := flag.String("c", pwd + "/wr.tsv", "keyword replace setting")
	
	// コマンドライン引数を切り離す
	flag.Parse()

	// 処理はディレクトリかファイルのどちらかのみ実施。両方指定されてた場合はディレクトリ優先
	if *dir != "" {
		fmt.Printf("loading [%v] directory.\n", *dir)
		exists := isExist(*dir)
		if exists {
			config := loadConfig(*setting)
			for l := range config {
				fmt.Println(config[l])
			}
			list := listFiles(*dir, *dir, *pattern)
			for i := range list {
				// 作業出力ディレクトリを作成
				pwd, _ := os.Getwd()
				outdir := pwd + "/" + strconv.FormatInt(time.Now().Unix(), 10)
				if err := os.Mkdir(outdir, 0777); err != nil {
					fmt.Println(err)
				}
				// 置換したファイルを保存
				name := list[i]
				ioutil.WriteFile(outdir+"/"+name, []byte(replace(*dir+"/"+name, config)), os.ModePerm)
			}
		} else {
			fmt.Printf("[%v] not exists.\n", *dir)
		}
	}
}

// isExist は指定されたパスがディレクトリであるか判定します。
func isExist(directoryname string) bool {
	defer func() {
		//err := recover()
		//fmt.Println(err)
	}()

	// 存在チェック
	info, err := os.Stat(directoryname)
	if err != nil {
		return false
	} else {
		// 対象がディレクトリか判定
		return info.IsDir()
	}
}

// listFiles は指定されたディレクトリ配下にあるファイルを再帰的に確認し pattern と一致する条件のファイルを取得します。
func listFiles(rootPath, searchPath, pattern string) (files []string) {
	fis, err := ioutil.ReadDir(searchPath)

	if err != nil {
		panic(err)
	}

	for _, fi := range fis {
		fullPath := filepath.Join(searchPath, fi.Name())

		if fi.IsDir() {
			files = append(files, listFiles(rootPath, fullPath, pattern)...)
		} else {
			// 対象ファイルのパターン指定がある場合は絞り込み
			if pattern != "" {
				match, _ := path.Match(pattern, fi.Name())
				if match == false {
					continue
				}
			}

			// 検索ディレクトリからの相対パス化
			rel, err := filepath.Rel(rootPath, fullPath)
			if err != nil {
				panic(err)
			}

			// 相対パスを覚えておく
			files = append(files, rel)
		}
	}

	return files
}

// loadFile はファイルの内容を読み込みます。
func loadFile(path string) string {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	desc := string(bs)
	return desc
}

// replace は指定された文字列を置換します。
func replace(path string, replacepattern []string) string {
	// ファイルを読み込む
	desc := loadFile(path)
	r := strings.NewReplacer(replacepattern...)
	replaced := r.Replace(desc)
	return replaced
}

// loadConfig は置換設定ファイルを読み込みます。
func loadConfig(setting string) (config []string) {
	pwd, _ := os.Getwd()
	fp, _ := os.Open(setting)

    reader := csv.NewReader(fp)
    reader.Comma = '\t'
    reader.LazyQuotes = true // ダブルクオートを厳密にチェックしない！
    for {
        record, err := reader.Read()
        if err == io.EOF {
        	break
        } else if err != nil {
            panic(err)
        }
        config = append(config, record...)
    }

    return config
}