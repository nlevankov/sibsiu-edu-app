## sibsiu-edu-app

![sibsiu](https://drive.google.com/uc?id=1AtFR_fJSV0hoyV2ctCjU71sJUwVHS_Vu)

A web application for recording grades and attendance info of students of the Siberian State Industrial University (SibSIU). In other words, it's like an electronic diary, but for the university.

## Deployment

Ready to work example (Windows 7 64bit):  
https://drive.google.com/uc?id=1wjmjDhwGo8h9CGc_oZbK9by3AOl8KfrY

Installation:
1. Compile the app with the `go` tool.
2. Install PostgreSQL server and create a database named "sibsiu_dev" (or with your custom name, but make sure the name is the same as in a .config file mentioned below).
3. Put assets, images, views (w/o *.go files) folders, .config file and the executable file of the app in the same folder.
4. It is ready to work. Launch the executable file and go to "localhost" in your browser (by default the port is set to 80).

If no configuration data is provided (.config), the app will start with default config data listed in the example below.
A .config file should be placed in the app's root folder.

Config example:

```
{
"port": 80,
"env": "dev",
"pepper": "secret-random-string", // authentication pepper string
"hmac_key": "secret-hmac-key", // this is necessary for generating remember tokens' hashes
"database": { // db connection info
"host": "localhost",
"port": 5432,
"user": "postgres",
"password": "123",
"name": "sibsiu_dev"
}
}
```

Also, the app can be launched with the `-prod` flag.
Provide this flag in production.
This ensures that a .config file is provided before the application starts.

Note: every time when the app starts the database schema is created with some test data. Every time this action drops all existed tables and data in the db.  
So, if you don't want that behavior, you have to comment this line in main.go and rebuild the app:
```
models.WithTesting()
```

## Description

There are two roles in the app atm:
* A student, who can only read her/his grades and attendance info.
* The group’s headman, who can read/write grades and attendance info of students in his/her group, make a report in MS Excel format with this information and change her/his group's status.

You have to sign in to work in the app. You can't sign up via web interface in the app due to the logic design, so you can use these test accounts to sign in (make sure the `models.WithTesting()` line is uncommented):
- levankov_nv (role: Student), pwd: 123123123
- dmitrieva_ag (role: The group’s headman) pwd: 123123123

## Notes

I made this app to: 
* understand common conceptions and idiomatic ways of developing web applications with Golang;
* learn the language itself;
* to figure out the MVC workflow;
* practice making CRUD operations and SQL queries with gorm or without it (raw SQL with parameterization);
* figure out how to develop private (internal) REST-like HTTP API with JSON as communication format;
* get my bachelor's degree.

This app is made according to the principles and ideas stated in the [Jon Calhoun's](https://twitter.com/joncalhoun) course ["Web Development With Go"](https://www.usegolang.com/). It uses many of the code provided in the course's book and it inherits the course's code license (MIT License).
Here is an excerpt from the book about the license:
```
All source code in the book is available under the MIT License. Put simply,
this means you can copy any of the code samples in this book and use them
on your own applications without owing me or anyone else money or anything
else. My only request is that you don’t use this source code to teach your own
course or write your own book.
```
Also, I used this great book:  
Brian W. Kernighan, Alan A. A. Donovan - The Go Programming Language (2015, Addison-Wesley Professional Computing).

This is a REST-like app, not REST-ful. It doesn't fully realize all the conceptions required by REST conventions.  It doesn't have caching in particular. But the architecture allows to implement these features later like layers, e.g. it allows to make every program layer independently and in similar style.

Error handling, package and function documentation, localization and testing are intentionally omitted almost everywhere in the app because I hadn't much time at the time of writing it (due to full-time education in the university) and I wasn't sure about whether I could implement the app at least at "it works" state or not. The app is not refactored.
I do realize that it is unmaintainable, but I wanted to focus on things listed above. So the app's alternative name is "TODO" :D  
It's my education project that gave me good experience in web programming with Golang and understanding of some basic conceptions, patterns and programming techniques in developing web apps with Golang.

Some of the used technologies/tools/packages:
* backend: 
    * Golang
    * Postgresql
    * gorm
    
* frontend: 
    * Twitter Bootstrap
    * jQuery  
    Yes I know that jQuery != JS, but I used jQuery bc it handles JSON well and provides handy AJAX API wrapper. 
    Also, I needed to use some custom multiple html selects that required some css/js code for the behavior that was required in my app.
    The solution was the jQuery plugin which implements those selects exactly as I needed.

* communication tools and formats:
    * REST
    * JSON  
    I considered choosing protobufs but wasn't sure I could learn and then integrate it in my app in time.
    * AJAX
    
* packages: 
    * github.com/gorilla/mux
    * net/http
    * net/url
    * encoding/json
    * context
    * github.com/jinzhu/gorm
    * github.com/gorilla/schema
    * github.com/360EntSecGroup-Skylar/excelize
    * crypto/hmac  
    // crypto/* packages are mainly used for generating hashes and auth salt strings
    * crypto/sha256
    * crypto/rand
    * encoding/base64
    * hash
    * golang.org/x/crypto/bcrypt
    * regexp
    * time
    * bytes
    * errors
    * github.com/gorilla/csrf
    * html/template
    * path/filepath

## Screenshots

![Screenshot 1](https://drive.google.com/uc?id=1hQOkTxMYyXbkCklOod4J-AKoEfQhEA-k)

![Screenshot 2](https://drive.google.com/uc?id=1bJVekCsUKg5GUBDeXJidCdSsvJTGMcVM)

![Screenshot 3](https://drive.google.com/uc?id=1YrdsW_pTRZDXNuWpKnCBnfHcPlK-QVXS)

![Screenshot 4](https://drive.google.com/uc?id=18SR2YX9_vITMBm8MBH3m9PQ3FHYvAcY1)

![Screenshot 5](https://drive.google.com/uc?id=1eS204IpGcGlgX4LV9Og73yCIxkis2UAI)

![Screenshot 6](https://drive.google.com/uc?id=17r5_j3YJDeWboz-XfDk_KEVwTMJL7yMM)

