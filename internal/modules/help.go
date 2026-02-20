/*
  - This file is part of YukkiMusic.
    *

  - YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
  - Copyright (C) 2025 TheTeamVivek
    *
  - This program is free software: you can redistribute it and/or modify
  - it under the terms of the GNU General Public License as published by
  - the Free Software Foundation, either version 3 of the License, or
  - (at your option) any later version.
    *
  - This program is distributed in the hope that it will be useful,
  - but WITHOUT ANY WARRANTY; without even the implied warranty of
  - MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  - GNU General Public License for more details.
    *
  - You should have received a copy of the GNU General Public License
  - along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package modules

import (
	"fmt"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
)

func init() {
	helpTexts["/help"] = fmt.Sprintf(`ℹ️ <b>Yardım Komutu</b>
<i>Genel bot yardımını veya belirli bir komut hakkında detaylı bilgiyi gösterir.</i>

<u>Kullanım:</u>
<code>/yardim</code> — Ana yardım menüsünü gösterir.  
<code>/yardim &lt;komut&gt;</code> — Belirli bir komut için yardım gösterir.

<b>💡 İpucu:</b> Herhangi bir komutun sonuna <code>-h</code> veya <code>--help</code> parametresi ekleyerek doğrudan o komutun yardımını alabilirsiniz, örn. <code>/oynat -h</code>

<b>⚠️ Not:</b> Bazı komutlar belirli bağlamlarla <b>sınırlandırılmıştır</b> (örneğin <b>Gruplar</b>, <b>Yöneticiler</b>, <b>Sudo Kullanıcıları</b> veya <b>Kurucu</b>).  
Eğer kısıtlı bir sohbette veya özel mesajda <code>-h</code> veya <code>--help</code> kullanmaya çalışırsanız, bot yanıt vermeyebilir.  
Bu komutların yardımını yine de görüntülemek için şu formatı kullanın:
<code>/yardim &lt;komut&gt;</code>

Daha fazla bilgi için <a href="%s">Destek Grubumuzu</a> ziyaret edin.`, config.SupportChat)

	// Türkçe komutu yardım menüsüne eşliyoruz
	helpTexts["/yardim"] = helpTexts["/help"]
}

func helpHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	if len(args) > 1 {
		cmd := args[1]
		if cmd != "pm_help" {
			if !strings.HasPrefix(cmd, "/") {
				cmd = "/" + cmd
			}
			return showHelpFor(m, cmd)
		}
	}

	if m.ChatType() != tg.EntityUser {
		m.Reply(
			F(m.ChannelID(), "help_private_only"),
			&tg.SendOptions{
				ReplyMarkup: core.GetGroupHelpKeyboard(m.ChannelID()),
			},
		)
		return tg.ErrEndGroup
	}

	m.Reply(
		F(m.ChannelID(), "help_main"),
		&tg.SendOptions{ReplyMarkup: core.GetHelpKeyboard(m.ChannelID())},
	)
	return tg.ErrEndGroup
}

func helpCB(c *tg.CallbackQuery) error {
	c.Edit(
		F(c.ChannelID(), "help_main"),
		&tg.SendOptions{ReplyMarkup: core.GetHelpKeyboard(c.ChannelID())},
	)
	c.Answer("")
	return tg.ErrEndGroup
}

func helpCallbackHandler(c *tg.CallbackQuery) error {
	data := c.DataString()
	c.Answer("")
	if data == "" {
		return tg.ErrEndGroup
	}
	chatID := c.ChannelID()
	parts := strings.SplitN(data, ":", 2)
	if len(parts) < 2 {
		return tg.ErrEndGroup
	}

	var text string
	btn := core.GetBackKeyboard(chatID)

	switch parts[1] {
	case "admins":
		text = F(chatID, "help_admin")
	case "sudoers":
		text = F(chatID, "help_sudo")
	case "owner":
		text = F(chatID, "help_owner")
	case "public":
		text = F(chatID, "help_public")
	case "main":
		text = F(chatID, "help_main")
		btn = core.GetHelpKeyboard(chatID)
	}

	c.Edit(text, &tg.SendOptions{ReplyMarkup: btn})
	return tg.ErrEndGroup
}
		switch parts[1] {
	case "admins":
		text = F(chatID, "help_admin")
	case "sudoers":
		text = F(chatID, "help_sudo")
	case "owner":
		text = F(chatID, "help_owner")
	case "public":
		text = F(chatID, "help_public")
	case "main":
		text = F(chatID, "help_main")
		btn = core.GetHelpKeyboard(chatID)
	}

	c.Edit(text, &tg.SendOptions{ReplyMarkup: btn})
	return tg.ErrEndGroup
}
