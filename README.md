# otpbase

An OTP code receiver/server/forwarder.

otpbase is a server which, in conjunction with Twilio,
can be used to receve OTP codes via SMS, store them for a brief period of
time, and forward them to another phone number.
The use case it was built for is ability to access the OTP codes
for applications and services which only offer the option to send the
codes via SMS messages (or proprietary applications) in a generic way.

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

## License

2 clause BSD
