/*
  - This file is part of YukkiMusic.
  - Edited by KIYICI BOSS (@officialkiyici)
*/
package modules

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/platforms"
	"main/internal/utils"
)

type playOpts struct {
	Force bool
	CPlay bool
	Video bool
}

const playMaxRetries = 3

func init() {
	helpTexts["/oynat"] = `<i>Sesli sohbette YouTube, Spotify veya diğer kaynaklardan müzik çalın.</i>

<u>Kullanım:</u>
<b>/oynat [sorgu/URL]</b> — Şarkı ara ve çal
<b>/oynat [yanıtlanan ses/video]</b> — Yanıtlanan medyayı çal

<b>🎵 Desteklenen Kaynaklar:</b>
• YouTube (videolar, oynatma listeleri)
• Spotify (şarkılar, albümler, oynatma listeleri)
• SoundCloud
• Doğrudan ses/video bağlantıları

<b>⚙️ Özellikler:</b>
• Sıra desteği - zaten çalıyorsa sona ekler
• Sesli sohbete otomatik katılım
• Süre sınırı kontrolü
• Çoklu parça desteği (oynatma listeleri)

<b>💡 Örnekler:</b>
<code>/oynat tarkan yolla</code>
<code>/oynat https://youtu.be/dQw4w9WgXcQ</code>

<b>⚠️ Notlar:</b>
• Botun sesli sohbette gerekli izinlere sahip olması gerekir
• Süre sınırını aşan parçalar atlanır
• Yaklaşan parçaları görmek için <code>/sira</code> kullanın
• Sırayı atlayıp anında çalmak için <code>/foynat</code> kullanın`

	helpTexts["/foynat"] = `<i>Sırayı atlayarak şarkıyı anında çalın.</i>

<u>Kullanım:</u>
<b>/foynat [sorgu/URL]</b> — Anında çalmaya başla
<b>/foynat [yanıtlanan ses/video]</b> — Yanıtlanan medyayı anında çal

<b>🎵 Davranış:</b>
• Mevcut çalınan şarkıyı durdurur
• Sırayı temizler
• Anında çalmaya başlar

<b>🔒 Kısıtlamalar:</b>
• Sadece <b>sohbet yöneticileri</b> veya <b>yetkili kullanıcılar</b> kullanabilir

<b>💡 Örnek:</b>
<code>/foynat acil duyuru müziği</code>`

	helpTexts["/voynat"] = `<i>Sesli sohbette video içeriği oynatın (video modu).</i>

<u>Kullanım:</u>
<b>/voynat [sorgu/URL]</b> — Videoyu oynat
<b>/voynat [yanıtlanan video]</b> — Yanıtlanan videoyu oynat

<b>📹 Özellikler:</b>
• Tam video oynatma desteği
• Ses + Video akışı
• Ses ile aynı sıra sistemi

<b>⚠️ Notlar:</b>
• Video akış izinleri gerektirir
• Zorla video oynatmak için <code>/fvoynat</code> kullanın`

	helpTexts["/fvoynat"] = `<i>Sırayı atlayarak video içeriğini anında oynatın.</i>

<u>Kullanım:</u>
<b>/fvoynat [sorgu/URL]</b> — Videoyu anında oynat

<b>🔒 Kısıtlamalar:</b>
• Sadece yönetici/yetkili komutudur`

	helpTexts["/koynat"] = `<i>Bağlı kanalın sesli sohbetinde çalın.</i>

<u>Kullanım:</u>
<b>/koynat [sorgu]</b> — Bağlı kanalda çal

<b>⚙️ Gerekli Kurulum:</b>
Önce şunu kullanın: <code>/kanaloynat --set [kanal_id]</code>`

	helpTexts["/kanaloynat"] = `<i>Kanal oynatma modu için bağlı kanalı yapılandırın.</i>

<u>Kullanım:</u>
<b>/kanaloynat --set [kanal_id]</b> — Bağlı kanalı ayarla

<b>⚙️ Davranış:</b>
• Bir kanalı mevcut gruba bağlar
• Bot kanala erişebilmelidir

<b>🔒 Kısıtlamalar:</b>
• Sadece <b>sohbet yöneticileri</b> yapılandırabilir`

	helpTexts["/oynatzorla"] = helpTexts["/foynat"]
	helpTexts["/fkoynat"] = helpTexts["/kfoynat"]
	helpTexts["/kvoynat"] = helpTexts["/vokoynat"]
}

func channelPlayHandler(m *tg.NewMessage) error {
	m.Reply(F(m.ChannelID(), "channel_play_depreciated"))
	return tg.ErrEndGroup
}

func playHandler(m *tg.NewMessage) error {
	return handlePlay(m, &playOpts{})
}

func fplayHandler(m *tg.NewMessage) error {
	return handlePlay(m, &playOpts{Force: true})
}

func cfplayHandler(m *tg.NewMessage) error {
	return handlePlay(m, &playOpts{Force: true, CPlay: true})
}

func vplayHandler(m *tg.NewMessage) error {
	return handlePlay(m, &playOpts{Video: true})
}

func fvplayHandler(m *tg.NewMessage) error {
	return handlePlay(m, &playOpts{Force: true, Video: true})
}

func vcplayHandler(m *tg.NewMessage) error {
	return handlePlay(m, &playOpts{CPlay: true, Video: true})
}

func fvcplayHandler(m *tg.NewMessage) error {
	return handlePlay(m, &playOpts{Force: true, CPlay: true, Video: true})
}

func cplayHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()

	if len(args) > 1 && args[1] == "--set" {
		if len(args) < 3 {
			m.Reply(
				F(chatID, "cplay_usage"),
				&tg.SendOptions{ParseMode: "HTML"},
			)
			return tg.ErrEndGroup
		}

		cplayIDStr := args[2]
		cplayID, err := strconv.ParseInt(cplayIDStr, 10, 64)
		if err != nil {
			m.Reply(
				F(chatID, "cplay_invalid_chat_id"),
				&tg.SendOptions{ParseMode: "HTML"},
			)
			return tg.ErrEndGroup
		}

		peer, err := m.Client.ResolvePeer(cplayID)
		if err != nil {
			m.Reply(
				F(chatID, "cplay_resolve_peer_fail"),
				&tg.SendOptions{ParseMode: "HTML"},
			)
			return tg.ErrEndGroup
		}

		chPeer, ok := peer.(*tg.InputPeerChannel)
		if !ok {
			m.Reply(
				F(chatID, "cplay_invalid_target"),
				&tg.SendOptions{ParseMode: "HTML"},
			)
			return tg.ErrEndGroup
		}

		fullChat, err := m.Client.ChannelsGetFullChannel(
			&tg.InputChannelObj{
				ChannelID:  chPeer.ChannelID,
				AccessHash: chPeer.AccessHash,
			},
		)
		if err != nil || fullChat == nil {
			gologging.ErrorF(
				"Failed to get full channel for cplay ID %d: %v",
				cplayID, err,
			)
			m.Reply(
				F(chatID, "cplay_channel_not_accessible"),
				&tg.SendOptions{ParseMode: "HTML"},
			)
			return tg.ErrEndGroup
		}

		if err := database.SetCPlayID(m.ChannelID(), cplayID); err != nil {
			gologging.ErrorF(
				"Failed to set cplay ID for chat %d: %v",
				m.ChannelID(), err,
			)
			m.Reply(
				F(chatID, "cplay_save_error"),
				&tg.SendOptions{ParseMode: "HTML"},
			)
			return err
		}

		m.Reply(
			F(chatID, "cplay_enabled", locales.Arg{
				"channel_id": cplayID,
			}),
			&tg.SendOptions{ParseMode: "HTML"},
		)
		return tg.ErrEndGroup
	}
	return handlePlay(m, &playOpts{CPlay: true})
}

func handlePlay(m *tg.NewMessage, opts *playOpts) error {
	mention := utils.MentionHTML(m.Sender)

	r, replyMsg, err := prepareRoomAndSearchMessage(m, opts.CPlay)
	if err != nil {
		return tg.ErrEndGroup
	}

	tracks, isActive, err := fetchTracksAndCheckStatus(
		m,
		replyMsg,
		r,
		opts.Video,
	)
	if err != nil {
		return tg.ErrEndGroup
	}

	tracks, availableSlots, err := filterAndTrimTracks(replyMsg, r, tracks)
	if err != nil {
		return tg.ErrEndGroup
	}

	if err := playTracksAndRespond(
		m, replyMsg, r, tracks, mention,
		isActive, opts.Force, availableSlots,
	); err != nil {
		return err
	}

	return tg.ErrEndGroup
}

func prepareRoomAndSearchMessage(
	m *tg.NewMessage,
	cplay bool,
) (*core.RoomState, *tg.NewMessage, error) {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return nil, nil, err
	}

	chatID := m.ChannelID()
	r.Parse()

	if len(r.Queue()) >= config.QueueLimit {
		m.Reply(F(chatID, "queue_limit_reached", locales.Arg{
			"limit": config.QueueLimit,
		}))
		return nil, nil, fmt.Errorf("queue limit reached")
	}

	parts := strings.SplitN(m.Text(), " ", 2)
	query := ""
	if len(parts) > 1 {
		query = strings.TrimSpace(parts[1])
	}

	if query == "" && !m.IsReply() {
		m.Reply(F(chatID, "no_song_query", locales.Arg{
			"cmd": getCommand(m),
		}))
		return nil, nil, fmt.Errorf("no song query")
	}

	// Searching messages
	searchStr := ""
	if query != "" {
		searchStr = F(chatID, "searching_query", locales.Arg{
			"query": html.EscapeString(query),
		})
	} else {
		searchStr = "**𝐌𝐞𝐯𝐳𝐮𝐲𝐮 𝐚𝐫𝐚𝐬̧𝐭𝐢𝐫𝐢𝐲𝐨𝐫𝐮𝐦, 𝐛𝐞𝐤𝐥𝐞 𝐛𝐞𝐛𝐞...**" // MUSTAFA ARAZ ŞİVESİ
	}

	replyMsg, err := m.Reply(searchStr)
	if err != nil {
		gologging.ErrorF("Failed to send searching message: %v", err)
		return nil, nil, err
	}

	return r, replyMsg, nil
}

func fetchTracksAndCheckStatus(
	m *tg.NewMessage,
	replyMsg *tg.NewMessage,
	r *core.RoomState,
	video bool,
) ([]*state.Track, bool, error) {
	tracks, err := safeGetTracks(m, replyMsg, m.ChannelID(), video)
	if err != nil {
		utils.EOR(replyMsg, err.Error())
		return nil, false, err
	}

	if len(tracks) == 0 {
		utils.EOR(replyMsg, F(m.ChannelID(), "no_song_found"))
		return nil, false, fmt.Errorf("no tracks found")
	}

	isActive := r.IsActiveChat()
	cs, err := core.GetChatState(r.ChatID())
	if err != nil {
		gologging.ErrorF("Error getting chat state: %v", err)
		utils.EOR(replyMsg, getErrorMessage(m.ChannelID(), err))
		return nil, false, err
	}

	activeVC, err := cs.IsActiveVC()
	if err != nil {
		gologging.ErrorF("Error checking voicechat state: %v", err)
		utils.EOR(replyMsg, getErrorMessage(m.ChannelID(), err))
		return nil, false, err
	}

	if !activeVC {
		utils.EOR(replyMsg, F(m.ChannelID(), "err_no_active_voicechat"))
		return nil, false, fmt.Errorf("no active voice chat")
	}

	banned, err := cs.IsAssistantBanned()
	if err != nil {
		gologging.ErrorF("Error checking assistant banned state: %v", err)
		utils.EOR(replyMsg, getErrorMessage(m.ChannelID(), err))
		return nil, false, err
	}

	if banned {
		utils.EOR(replyMsg,
			F(m.ChannelID(), "err_assistant_banned", locales.Arg{
				"user": utils.MentionHTML(cs.Assistant.User),
				"id":   utils.IntToStr(cs.Assistant.User.ID),
			}),
		)
		return nil, false, fmt.Errorf("assistant banned")
	}

	present, err := cs.IsAssistantPresent()
	if err != nil {
		gologging.ErrorF("Error checking assistant presence: %v", err)
		utils.EOR(replyMsg, getErrorMessage(m.ChannelID(), err))
		return nil, false, err
	}

	if !present {
		if err := cs.TryJoin(); err != nil {
			gologging.ErrorF("Error joining assistant: %v", err)
			utils.EOR(replyMsg, getErrorMessage(m.ChannelID(), err))
			return nil, false, err
		}
		time.Sleep(1 * time.Second)
	}
	return tracks, isActive, nil
}

func filterAndTrimTracks(
	replyMsg *tg.NewMessage,
	r *core.RoomState,
	tracks []*state.Track,
) ([]*state.Track, int, error) {
	chatID := replyMsg.ChannelID()

	var filteredTracks []*state.Track
	var skippedTracks []string

	for _, track := range tracks {
		if track.Duration > config.DurationLimit {
			skippedTracks = append(
				skippedTracks,
				html.EscapeString(utils.ShortTitle(track.Title, 35)),
			)
			continue
		}
		filteredTracks = append(filteredTracks, track)
	}

	// Some tracks were skipped due to duration limit
	if len(skippedTracks) > 0 {

		// CASE 1: Only one track and it was skipped
		if len(tracks) == 1 && len(filteredTracks) == 0 {
			utils.EOR(
				replyMsg,
				F(chatID, "play_single_track_too_long", locales.Arg{
					"limit_mins": formatDuration(config.DurationLimit),
					"title":      skippedTracks[0],
				}),
			)
			return nil, 0, fmt.Errorf("single long track skipped")
		}

		// CASE 2: Multiple tracks skipped
		var b strings.Builder

		b.WriteString(
			F(chatID, "play_multiple_tracks_too_long_header", locales.Arg{
				"count":      len(skippedTracks),
				"limit_mins": config.DurationLimit / 60,
			}),
		)
		b.WriteString("\n")

		for i, title := range skippedTracks {
			if i < 5 {
				b.WriteString(
					F(chatID, "play_multiple_tracks_too_long_item", locales.Arg{
						"title": title,
					}) + "\n",
				)
			} else {
				b.WriteString(F(chatID, "play_multiple_tracks_too_long_more", locales.Arg{
					"remaining": len(skippedTracks) - i,
				}) + "\n")
				break
			}
		}

		utils.EOR(replyMsg, b.String())
		time.Sleep(1 * time.Second)
	}

	// Keep only accepted tracks
	tracks = filteredTracks

	// CASE: everything was skipped
	if len(tracks) == 0 {
		utils.EOR(replyMsg, F(chatID, "play_all_tracks_skipped"))
		return nil, 0, fmt.Errorf("all tracks skipped")
	}

	// Respect queue limit
	availableSlots := config.QueueLimit - len(r.Queue())
	if availableSlots < len(tracks) {
		tracks = tracks[:availableSlots]
		gologging.WarnF(
			"Queue full — adding only %d tracks out of requested.",
			availableSlots,
		)
	}

	return tracks, availableSlots, nil
}

func playTracksAndRespond(
	m *tg.NewMessage,
	replyMsg *tg.NewMessage,
	r *core.RoomState,
	tracks []*state.Track,
	mention string,
	isActive, force bool,
	availableSlots int,
) error {
	chatID := m.ChannelID()

	for i, track := range tracks {
		track.Requester = mention
		title := html.EscapeString(utils.ShortTitle(track.Title, 25))
		var filePath string

		// Download first track if needed
		if i == 0 && (!isActive || force) {
			var opt *tg.SendOptions
			if track.Duration > 420 {
				opt = &tg.SendOptions{
					ReplyMarkup: core.GetCancelKeyboard(chatID),
				}
			}

			downloadingText := F(chatID, "play_downloading_song", locales.Arg{
				"title": title,
			})
			replyMsg, _ = utils.EOR(replyMsg, downloadingText, opt)

			ctx, cancel := context.WithCancel(context.Background())
			downloadCancels[m.ChannelID()] = cancel
			defer func() {
				if _, ok := downloadCancels[m.ChannelID()]; ok {
					delete(downloadCancels, m.ChannelID())
					cancel()
				}
			}()

			path, err := safeDownload(ctx, track, replyMsg, chatID)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					utils.EOR(
						replyMsg,
						F(chatID, "play_download_canceled", locales.Arg{
							"user": mention,
						}),
					)
				} else {
					utils.EOR(replyMsg, F(chatID, "play_download_failed", locales.Arg{
						"title": title,
						"error": html.EscapeString(err.Error()),
					}))
				}
				return tg.ErrEndGroup
			}

			filePath = path
			gologging.InfoF("Downloaded track to %s", filePath)
		}

		// 🔁 play with retry
		if err := playTrackWithRetry(r, track, filePath, force && i == 0, replyMsg); err != nil {
			return err
		}
		r.DeleteData("rec_cache")
		sendPlayLogs(m, track, (isActive && !force) || i > 0)
	}

	mainTrack := tracks[0]

	// ---------- Now Playing / Added to queue ----------
	if !isActive || (force && len(tracks) > 0) {
		title := html.EscapeString(utils.ShortTitle(mainTrack.Title, 25))
		btn := core.GetPlayMarkup(chatID, r, false)

		opt := &tg.SendOptions{
			ParseMode:   "HTML",
			ReplyMarkup: btn,
		}

		if mainTrack.Artwork != "" && shouldShowThumb(chatID) {
			opt.Media = utils.CleanURL(mainTrack.Artwork)
		}

		nowPlayingText := F(chatID, "stream_now_playing", locales.Arg{
			"url":      mainTrack.URL,
			"title":    title,
			"duration": formatDuration(mainTrack.Duration),
			"by":       mention,
		})

		replyMsg, _ = utils.EOR(replyMsg, nowPlayingText, opt)
		r.SetMystic(replyMsg)

		if len(tracks) > 1 {
			addedCount := len(tracks) - 1

			var b strings.Builder
			b.WriteString(F(chatID, "play_added_multiple_header", locales.Arg{
				"count": addedCount,
				"user":  mention,
			}))
			b.WriteString("\n\n")

			if availableSlots <= len(tracks) {
				b.WriteString(F(chatID, "play_queue_limit_hint"))
				b.WriteString("\n")
			}

			b.WriteString(F(chatID, "play_queue_view_hint"))
			replyMsg.Respond(b.String())
		}
	} else {
		if len(tracks) == 1 {
			title := html.EscapeString(utils.ShortTitle(mainTrack.Title, 25))
			btn := core.GetPlayMarkup(chatID, r, true)
			opt := &tg.SendOptions{
				ParseMode:   "HTML",
				ReplyMarkup: btn,
			}
			if mainTrack.Artwork != "" && shouldShowThumb(chatID) {
				opt.Media = utils.CleanURL(mainTrack.Artwork)
			}

			addedText := F(chatID, "play_added_to_queue_single", locales.Arg{
				"url":      mainTrack.URL,
				"title":    title,
				"duration": formatDuration(mainTrack.Duration),
				"by":       mention,
			})

			utils.EOR(replyMsg, addedText, opt)
		} else {
			var b strings.Builder
			b.WriteString(F(chatID, "play_added_multiple_header", locales.Arg{
				"count": len(tracks),
				"user":  mention,
			}))
			b.WriteString("\n\n")

			if availableSlots <= len(tracks) {
				b.WriteString(F(chatID, "play_queue_limit_hint"))
				b.WriteString("\n")
			}

			b.WriteString(F(chatID, "play_queue_view_hint"))
			utils.EOR(replyMsg, b.String())
		}
	}

	return nil
}

func playTrackWithRetry(
	r *core.RoomState,
	track *state.Track,
	filePath string,
	force bool,
	replyMsg *tg.NewMessage,
) error {
	for attempt := 1; attempt <= playMaxRetries; attempt++ {

		if r.Destroyed() {
			gologging.Info("Room destroyed during retry, aborting")
			replyMsg.Delete()
			return tg.ErrEndGroup
		}

		err := r.Play(track, filePath, force)
		if err == nil {
			if attempt > 1 {
				gologging.Info(
					"Successfully played after retry attempt " + utils.IntToStr(
						attempt,
					),
				)
			}
			return nil
		}

		// FloodWait
		if wait := tg.GetFloodWait(err); wait > 0 {
			gologging.Error(
				"FloodWait detected (" + strconv.Itoa(
					wait,
				) + "s). Retrying... (attempt " + utils.IntToStr(
					attempt,
				) + ")",
			)
			time.Sleep(time.Duration(wait) * time.Second)
			continue
		}

		if strings.Contains(
			err.Error(),
			"Streaming is not supported when using RTMP",
		) {
			utils.EOR(
				replyMsg,
				F(replyMsg.ChannelID(), "rtmp_streaming_not_supported"),
			)
			core.DeleteRoom(r.ChatID())
			return tg.ErrEndGroup
		}

		if strings.Contains(err.Error(), "group call") &&
			strings.Contains(err.Error(), "is closed") {
			utils.EOR(
				replyMsg,
				F(replyMsg.ChannelID(), "err_no_active_voicechat"),
			)
			return tg.ErrEndGroup
		}

		if tg.MatchError(err, "GROUPCALL_INVALID") {
			gologging.Error("GROUPCALL_INVALID err occurred. Returning...")
			core.DeleteRoom(r.ChatID())
			utils.EOR(replyMsg, F(replyMsg.ChannelID(), "play_unable"))
			return tg.ErrEndGroup
		}

		// INTERDC_X_CALL_ERROR → retry
		if tg.MatchError(err, "INTERDC_X_CALL_ERROR") {
			gologging.Error(
				"INTERDC_X_CALL_ERROR occurred. Retrying... (attempt " + utils.IntToStr(
					attempt,
				) + ")",
			)
			time.Sleep(2 * time.Second)
			continue
		}

		// Last attempt failed
		if attempt == playMaxRetries {
			gologging.Error(
				"❌ Failed to play after " + utils.IntToStr(
					playMaxRetries,
				) + " attempts. Error: " + err.Error(),
			)
			utils.EOR(
				replyMsg,
				F(
					replyMsg.ChannelID(),
					"play_failed",
					locales.Arg{"error": err.Error()},
				),
			)
			return err
		}

		gologging.Error(
			"Unexpected error occurred. Retrying... (attempt " + utils.IntToStr(
				attempt,
			) + "): " + err.Error(),
		)
	}

	return nil
}

type msgFn func(chatID int64, err error) string

var errMessageMap = map[error]msgFn{
	core.ErrAdminPermissionRequired: func(chatID int64, _ error) string {
		return F(chatID, "err_admin_permission_required")
	},
	core.ErrAssistantGetFailed: func(chatID int64, e error) string {
		gologging.Error(e)
		return F(chatID, "err_assistant_get_failed", locales.Arg{
			"error": e.Error(),
		})
	},
	core.ErrAssistantJoinRateLimited: func(chatID int64, _ error) string {
		return F(chatID, "err_assistant_join_rate_limited")
	},

	core.ErrAssistantJoinRequestSent: func(chatID int64, _ error) string {
		return F(chatID, "err_assistant_join_request_sent")
	},

	core.ErrAssistantInviteLinkFetch: func(chatID int64, e error) string {
		return F(chatID, "err_assistant_invite_link_fetch", locales.Arg{
			"error": e.Error(),
		})
	},

	core.ErrAssistantInviteFailed: func(chatID int64, e error) string {
		return F(chatID, "err_assistant_invite_failed", locales.Arg{
			"error": e.Error(),
		})
	},

	core.ErrFetchFailed: func(chatID int64, e error) string {
		return F(chatID, "err_fetch_failed", locales.Arg{
			"error": e.Error(),
		})
	},

	core.ErrPeerResolveFailed: func(chatID int64, _ error) string {
		return F(chatID, "err_peer_resolve_failed")
	},
}

func getErrorMessage(chatID int64, err error) string {
	if err == nil {
		return ""
	}

	for key, fn := range errMessageMap {
		if errors.Is(err, key) {
			return fn(chatID, err)
		}
	}

	return F(chatID, "err_unknown", locales.Arg{
		"error": err.Error(),
	})
}

// KRİTİK GÜNCELLEME BURADA: MESSAGE_IDS_EMPTY hatasını engelleyen zırh!
func safeGetTracks(
	m, replyMsg *tg.NewMessage,
	chatID int64,
	video bool,
) (tracks []*state.Track, err error) {
	defer func() {
		if r := recover(); r != nil {
			utils.EOR(replyMsg, F(chatID, "err_fetch_tracks"))
			panic(r)
		}
	}()

	tracks, err = platforms.GetTracks(m, video)

	// Eğer Telegram boş ID gönderdin diyorsa araya Kıyıcı Boss girer!
	if err != nil && strings.Contains(err.Error(), "MESSAGE_IDS_EMPTY") {
		return nil, errors.New("**𝐋𝐚 𝐠𝐚𝐫𝐝𝐚𝐬̧, 𝐯𝐢𝐝𝐞𝐨 𝐛𝐨𝐳𝐮𝐤 𝐯𝐞𝐲𝐚 𝐬𝐢𝐥𝐢𝐧𝐦𝐢𝐬̧ (𝐌𝐄𝐒𝐒𝐀𝐆𝐄_𝐈𝐃𝐒_𝐄𝐌𝐏𝐓𝐘). 𝐁𝐚𝐬̧𝐤𝐚 𝐛𝐢𝐫 𝐯𝐢𝐝𝐞𝐨 𝐝𝐞𝐧𝐞 𝐛𝐞𝐛𝐞!**")
	}

	return tracks, err
}

func safeDownload(
	ctx context.Context,
	track *state.Track,
	replyMsg *tg.NewMessage,
	chatID int64,
) (path string, err error) {
	defer func() {
		if r := recover(); r != nil {
			utils.EOR(replyMsg, F(chatID, "err_download_internal"))
			panic(r)
		}
	}()

	path, err = platforms.Download(ctx, track, replyMsg)
	return path, err
}
