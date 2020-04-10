package main

import (
    "bytes"
    "crypto/subtle"
	"fmt"
    "io"
//    "io/ioutil"
    "log"
	"net/http"
    "os"
    "path"
    "path/filepath"
    "strings"
//    "strconv"
    "time"
)

const
(
    maxUploadSize = 100 * 1024 * 1024 + 512 // 100mb plus fringe
    ADMIN_USER = "admin"
    ADMIN_PASSWORD = "admin"
    USER = "user"
    PASSWORD = "user"
)

func BasicAuth(handler http.HandlerFunc, realm string, account string) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    var cred_user string
    var cred_pass string

    if account == "admin" {
        cred_user = ADMIN_USER
        cred_pass = ADMIN_PASSWORD
    } else {
        cred_user = USER
        cred_pass = PASSWORD
    }

    user, pass, ok := r.BasicAuth()
    if ( !ok ||
      subtle.ConstantTimeCompare([]byte(user), []byte(cred_user)) != 1 ||
      subtle.ConstantTimeCompare([]byte(pass), []byte(cred_pass)) != 1 ) {
        w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
        w.WriteHeader(401)
        w.Write([]byte("You are Unauthorized for access.\n"))
        return
    }
    handler(w, r)
    }
}

func check(err error) {
    if err != nil {
        //panic(err)
        fmt.Println("ERROR:", err)
    }
}

func index(w http.ResponseWriter, r *http.Request) {
    fullpath := r.URL.Path
    filepath := "static/upload" + fullpath

    switch r.Method {
    case "GET":
        if fullpath != "/" {
            http.ServeFile(w, r, filepath)
        } else {
            http.ServeFile(w, r, "static/index.html")
        }
    case "PUT":
        fallthrough
    case "POST":
        // Restrict the whole request body to 10kb
        r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

        if r.ContentLength > maxUploadSize {
            http.Error(w, fmt.Sprintf("Sorry, %s was too large. The limit is %d bytes, your request was %d bytes.", r.Method, maxUploadSize, r.ContentLength), http.StatusRequestEntityTooLarge)
			return
		}

        if err := r.ParseMultipartForm(maxUploadSize); err != nil {
            // if not form, then use full body as upload
            upload := r.Body
            storeFile(filepath, upload, w, r)
            //http.ServeFile(w, r, filepath)
		} else {
            for key, _ := range r.MultipartForm.File {
                upload, _, err := r.FormFile(key)
                if err != nil {
                    http.Error(w, fmt.Sprintf("Sorry, %s was wrong (input form %s has to be of type Multipart)", r.Method, key), http.StatusBadRequest)
                    return
                }
                storeFile(filepath, upload, w, r)
                //http.ServeFile(w, r, filepath)
            }
        }
    default:
        fmt.Fprintf(w, "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
        fmt.Fprintf(w, "<html>\n<head>\n<title>Failed</title>\n</head>\n<body>\n")
        fmt.Fprintf(w, "Sorry, only GET, PUT and POST methods are supported.")
        fmt.Fprintf(w, "</body>\n</html>\n")
    }
}

func storeFile(filepath string,  upload io.ReadCloser, w http.ResponseWriter, r *http.Request) {
    filedir := path.Dir(filepath)
    title := "Success"
    message := fmt.Sprintf("Successfully uploaded File %s, <a href=\"%s\">link</a>", r.URL.Path, r.URL.Path)
    fmt.Println("STORE:", filepath)

    err := os.MkdirAll(filedir, 0777)
    if err != nil {
        fmt.Println("ERROR::MkdirAll", err)
        title = "Mkdir failed"
        message = fmt.Sprintf("Could not create folder structure (probably an existing file named like one of your folders)")
    } else {
        f, err := os.Create(filepath)
        if err != nil {
            fmt.Println("ERROR::Create", err)
            title = "File creation failed"
            message = fmt.Sprintf("Could not create file")
        } else {
            _, err = io.Copy(f, upload)
            if err != nil {
                fmt.Println("ERROR::Copy", err)
                title = "Copy of upload content failed"
                message = fmt.Sprintf("Could not copy uploaded content to new file")
            } else {
                err = f.Close()
                if err != nil {
                    fmt.Println("ERROR::Close", err)
                    title = "Closing of new file failed"
                    message = fmt.Sprintf("Could not close the newly created file")
                }
            }
        }
    }
    fmt.Fprintf(w, "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
    fmt.Fprintf(w, "<title>%s</title>\n</head>\n<body>\n", title)
    fmt.Fprintf(w, "%s\n", message)
    fmt.Fprintf(w, "</body>\n</html>\n")
}

func robot(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "static/robots.txt")
}

func sitemap(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "static/sitemap.xml")
}

func favicon(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "static/favicon.ico")
}

func deleteFile(path string) (string, error) {
    //trim whitespaces
    path = strings.Trim(path, " ")
    //remove relative path
    //add local path prefix to contain to upload subfolder
    path = "static/upload/" + path
    path = strings.Replace(path, "/./", "/", -1)
    path = strings.Replace(path, "/../", "/", -1)

    info, err := os.Stat(path)
    if os.IsNotExist(err) {
        return fmt.Sprintf("File %s does not exist\n", path), err
    }

    if ( ! info.IsDir() ) {
        err = os.Remove(path)
        if err != nil {
            return fmt.Sprintf("Deletion of %s failed.\n",path), err
        }
        return fmt.Sprintf("Deleted: %s (%d bytes)\n", path, info.Size()), nil
    } else {
        return fmt.Sprintf("Folder %s cannot be deleted.\n",path), err
    }
}

func deletehandler(w http.ResponseWriter, r *http.Request) {
    keys, ok := r.URL.Query()["file"]

    if !ok || len(keys[0]) < 1 {
        fmt.Fprintf(w, "Url Param 'file' is missing")
        return
    }

    // Query()["key"] will return an array of items, 
    // we only want the single item.
    path := keys[0]
    text, err := deleteFile(path)

    fmt.Fprintf(w, "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
    fmt.Fprintf(w, "    <title>Delete File</title>\n</head>\n<body>\n")

    if err != nil {
        fmt.Fprintf(w, "    Deletion failed: %s\n", err)
    } else {
        fmt.Fprintf(w, "    %s\n", text)
    }

    fmt.Fprintf(w, "</body>\n</html>\n")
}

func IsDirEmpty(name string) (bool, error) {
         f, err := os.Open(name)
         if err != nil {
                 return false, err
         }
         defer f.Close()

         // read in ONLY one file
         _, err = f.Readdir(1)

         // and if the file is EOF... well, the dir is empty.
         if err == io.EOF {
                 return true, nil
         }
         return false, err
 }

func deleteExpired() (string, error) {
    var buf bytes.Buffer
    expired := time.Now().Add(- time.Hour * 24)

    err := filepath.Walk("static/upload/", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if ( ! info.IsDir() ) {
            if info.ModTime().Before(expired) {
                err = os.Remove(path)
                if err != nil {
                    return err
                }
                buf.WriteString(fmt.Sprintf("Deleted: %s (%d bytes)\n", path, info.Size()))
            }
        } else {
            if (path != "static/upload/") {
                isempty, err := IsDirEmpty(path)
                check(err)
                if (isempty) {
                    err = os.Remove(path)
                    if err != nil {
                        buf.WriteString(fmt.Sprintf("Could not delete folder: %s (%d bytes), err: %s\n", path, info.Size(), err))
                    } else {
                        buf.WriteString(fmt.Sprintf("Deleted Folder: %s (%d bytes)\n", path, info.Size()))
                    }
                }
            }
        }
        return nil
    })
    if err != nil {
        return buf.String(), err
    }

    return buf.String(), nil
}

func status(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
    fmt.Fprintf(w, "    <title>pieni status</title>\n")
    css :=`    <style>
    table, th, td {
      border: 1px solid black;
      border-collapse: collapse;
    }
    th, td {
      padding: 5px;
    }
    th {
      text-align: left;
    }
    tr:nth-child(even) {background: #CCC}
    tr:nth-child(odd) {background: #FFF}
    </style>`
    fmt.Fprintf(w, "%s\n</head>\n", css)
    fmt.Fprintf(w, "<body>\n<pre>")
	fmt.Fprintf(w, "Uploaded Files\n")
	fmt.Fprintf(w, "<table>\n")
    fmt.Fprintf(w, "<tr><th>Name</th><th>Modtime</th><th>Size</th><th>Delete</th></tr>\n")

    err := filepath.Walk("static/upload/", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if (! info.IsDir()) {
            fmt.Fprintf(w, "<tr><td><a href=\"%s%s%s\">%s</a></td><td>%s</td><td>%s</td>",
                r.URL.Scheme, r.URL.Host, strings.TrimPrefix(path, "static/upload/"), strings.TrimPrefix(path, "static/upload/"), info.ModTime(), ByteCountBinary(info.Size()))
            fmt.Fprintf(w, "<td><a href=\"delete?file=%s\">delete</a></td></tr>\n",strings.TrimPrefix(path, "static/upload/"))
        }
        return nil
    })
    if err != nil {
        fmt.Fprintf(w, "Error listing files: %v\n", err)
    }

    fmt.Fprintf(w, "</table></pre>")
    deletedFiles, _ := deleteExpired()
    fmt.Fprintf(w, "<pre>\n%s</pre>\n", deletedFiles)
    fmt.Fprintf(w, "</body>\n</html>\n")
}

func ByteCountBinary(b int64) string {
        const unit = 1024
        if b < unit {
                return fmt.Sprintf("%d B", b)
        }
        div, exp := int64(unit), 0
        for n := b / unit; n >= unit; n /= unit {
                div *= unit
                exp++
        }
        return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func main() {
    //make sure the upload folder is there
    err := os.MkdirAll("static/upload",0777)
    check(err)

    http.HandleFunc("/", BasicAuth(index, "Please enter your username and password", "user"))
    http.HandleFunc("/robots.txt", robot)
    http.HandleFunc("/sitemap.xml", sitemap)
	http.HandleFunc("/status", BasicAuth(status, "Please enter administrative credentials", "admin"))
	http.HandleFunc("/delete", BasicAuth(deletehandler, "Please enter administrative credentials", "admin"))
    http.HandleFunc("/favicon.ico", favicon)

    // create a ticker to run the deleteExpired every hour
    ticker := time.NewTicker(1 * time.Hour)
    go func() {
        for _ = range ticker.C {
            deleteExpired()
            fmt.Println("Ticker: check for expired files and empty folders")
        }
    }()

    port, ok := os.LookupEnv("PIENI_PORT")
    if !ok {
        port = ":3001"
    } else {
        port = ":" + port
    }
	log.Fatal(http.ListenAndServe(port, nil))
    ticker.Stop()
}

