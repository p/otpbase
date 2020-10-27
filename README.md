# otpbase

An OTP code receiver/server/forwarder.

otpbase is a server which, in conjunction with Twilio,
can be used to receve OTP codes via SMS, store them for a brief period of
time, and forward them to another phone number.
The use case it was built for is ability to access the OTP codes
for applications and services which only offer the option to send the
codes via SMS messages (or proprietary applications) in a generic way.

otpbase can also generate TOTP codes. This functionality is similar to what
Google Authenticator provides.

In order to use otpbase, the following setup is necessary:

1. A [Twilio](https://www.twilio.com/) account.
2. A phone number in Twilio. These cost money, however a US phone number
with SMS support costs $1/month at the time of this writing.
There are also per-SMS fees, though again at the time of this writing
they are under 1 cent/message for US phone numbers.
3. Configure Twilio to invoke otpbase as a web hook via POST.
4. If desired, configure otpbase to forward the messages to another
phone number.

## Installation

otpbase's dependencies are managed with [go-get-deps](https://github.com/p/go-get-deps).

## Configuration

otpbase recognizes the following environment variables at runtime:

- DEBUG: enable Gin debug mode
- DB_PATH: Path to database file for storing TOTP application tokens
- PORT: port number to bind to (default is 8092)
- HTTP_USER: enable HTTP authentication for retrieving OTP codes, specify
  the user name
- HTTP_PASSWORD: enable HTTP authentication for retrieving OTP codes, specify
  the password

## Usage

To view codes received from SMSes: http://localhost

To add an application for generating OTP codes:

    curl -d 'name=myapp&secret=OTPSECRET' http://localhost/apps

Then to retrieve an OTP code: http://localhost/apps/myapp

## Caveats

When I gave a Twilio SMS number to Google for MFA, Google accpted the number
but never sent any messages to it. I imagine Google knows whether each phone
number in existence is attached to a physical phone or not, or at the very
least has this as one of their goals. This makes the SMS part of otpbase
unusable for Google MFA. Fortunately it is possible to set up otpbase
in lieu of using Google Authenticator, which Google so far accepts.

## License

2 clause BSD
