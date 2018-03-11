package centralaccount

import (
	"bitbucket.org/0xor1/task/server/util/core"
	"fmt"
)

func emailSendMultipleAccountPolicyNotice(ctx *core.Ctx, address string) {
	ctx.MailClient().Send([]string{address}, "sendMultipleAccountPolicyNotice")
}

func emailSendActivationLink(ctx *core.Ctx, address, activationCode string) {
	ctx.MailClient().Send([]string{address}, fmt.Sprintf("sendActivationLink: activationCode: %s", activationCode))
}

func emailSendPwdResetLink(ctx *core.Ctx, address, resetCode string) {
	ctx.MailClient().Send([]string{address}, fmt.Sprintf("sendPwdResetLink: resetCode: %s", resetCode))
}

func emailSendNewEmailConfirmationLink(ctx *core.Ctx, currentAddress, newAddress, confirmationCode string) {
	ctx.MailClient().Send([]string{newAddress}, fmt.Sprintf("sendNewEmailConfirmationLink: currentAddress: %s newAddress: %s confirmationCode: %s", currentAddress, newAddress, confirmationCode))
}
