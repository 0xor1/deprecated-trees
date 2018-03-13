package centralaccount

import (
	"bitbucket.org/0xor1/task/server/util/ctx"
	"fmt"
)

func emailSendMultipleAccountPolicyNotice(ctx ctx.Ctx, address string) {
	ctx.MailClient().Send([]string{address}, "sendMultipleAccountPolicyNotice")
}

func emailSendActivationLink(ctx ctx.Ctx, address, activationCode string) {
	ctx.MailClient().Send([]string{address}, fmt.Sprintf("sendActivationLink: activationCode: %s", activationCode))
}

func emailSendPwdResetLink(ctx ctx.Ctx, address, resetCode string) {
	ctx.MailClient().Send([]string{address}, fmt.Sprintf("sendPwdResetLink: resetCode: %s", resetCode))
}

func emailSendNewEmailConfirmationLink(ctx ctx.Ctx, currentAddress, newAddress, confirmationCode string) {
	ctx.MailClient().Send([]string{newAddress}, fmt.Sprintf("sendNewEmailConfirmationLink: currentAddress: %s newAddress: %s confirmationCode: %s", currentAddress, newAddress, confirmationCode))
}