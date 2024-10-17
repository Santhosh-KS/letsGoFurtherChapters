package main

import (
	"fmt"
	// "github.com/go-mail/mail/v2"
	"smtpMails.vbuntu.org/internal/mailer"
)

func main() {
	fmt.Println("Hello world!")
	host := "box.masters.in" // Give valid mail server here
	port := 465
	username := "me@examplel.com" // Give valid email address to be used to send the emails
	password := "********"        // Give valid password here.
	myMailer := mailer.New(host, port, username, password, username)
	type User struct {
		name string
		id   int
	}
	user := User{"John Doe", 1}
	err := myMailer.Send("a@a.com", "user_welcome.tmpl", user) // --> Change the address
	if err != nil {
		fmt.Println("Error occourd ", err)
	}
	/* m := mail.NewMessage()
	// m.SetHeader("From", "me@myemail.com")
	m.SetHeader("From", "me@example.com")
	// m.SetHeader("From", "test@gmail.com")
	m.SetHeader("To", "santhosh@vbuntu.in", "test@gmail.com")
	m.SetHeader("Subject", "Hello!")
	body := "Hello <b>Bob</b> and <i>Cora</i>!"
	mime := "MIME-version:1.0; \n content-type: text/html; charset=\"UTF-8\";\n\n"
	m.SetBody("text/html", mime+body)
	// m.Attach("/home/statemate/Item1.jpg")

	d := mail.NewDialer("box.example.com", 465, "me@mailmasters.in", "****")
	// d := mail.NewDialer("smtp.office365.com", 587, "me@myemail.com", "*****")
	// d := mail.NewDialer("smtp.gmail.com", 587, "test@gmail.com", "****")
	d.StartTLSPolicy = mail.MandatoryStartTLS

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	} */
}
