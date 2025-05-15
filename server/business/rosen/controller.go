package rosen

import (
	"os"

	"github.com/GoROSEN/go-circle-web3-service/devwallets"
	"github.com/GoROSEN/rosen-apiserver/core/common"
	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/GoROSEN/rosen-apiserver/core/event"
	"github.com/GoROSEN/rosen-apiserver/features/account"
	"github.com/GoROSEN/rosen-apiserver/features/member"
	"github.com/GoROSEN/rosen-apiserver/features/message"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/google/martian/log"
	"github.com/oschwald/geoip2-golang"
	"gorm.io/gorm"
)

// Controller 控制器
type Controller struct {
	common.CrudController
	common.OssController
	message.MsgMod

	service        *Service
	memberService  *member.Service
	accountService *account.AccountService
}

// NewController 初始化控制器
func NewController(r *gin.Engine, db *gorm.DB, rds *redis.Client, geoip *geoip2.Reader) *Controller {

	c := &Controller{}
	c.SetupCrud(db)
	c.SetupOSS("rosen/")
	c.SetupMsgMod(db, rds)
	c.accountService = account.NewAccountService(db)
	c.memberService = member.NewService(db, rds)

	cdevcfg := config.GetConfig().Rosen.CircleService
	var cdevwallet *devwallets.CircleDevWalletsService
	if cdevcfg.Enable {
		log.Infof("initializing circle web services...")
		pubkey, _ := os.ReadFile(cdevcfg.DevWallet.PublicKey)
		cdevwallet = devwallets.NewCircleDevWalletsService(cdevcfg.Host, cdevcfg.Apikey, string(pubkey), cdevcfg.DevWallet.Secret)
	}
	c.service = NewService(db, rds, c.accountService, geoip, &c.MsgMod, cdevwallet)

	open := r.Group("/api/open/rosen")
	c.setupOpenMemberController(open)
	c.setupOpenAlphaController(open)

	event.GetPublisher().AddListener(c)

	return c
}

// event listener
func (c *Controller) HandleEvent(event string, data interface{}) {

	if event == member.PostLoginEvent {
		// 登录后处理
		m := data.(*member.Member)
		var extra MemberExtra
		if err := c.Crud.FindModelWhere(&extra, "member_id = ?", m.ID); err != nil {
			log.Errorf("cannot get member by token: %v", err)
			return
		}
		if len(extra.ChatTranslationLang) == 0 && len(m.Language) > 0 {
			extra.ChatTranslationLang = m.Language
			extra.EnableChatTranslation = true
			if err := c.service.UpdateModel(&extra, []string{"chat_translation_lang", "enable_chat_translation"}, nil); err != nil {
				log.Errorf("cannot update chat_translation_lang: %v", err)
			}
		}
	} else if event == member.PostBlockUserEvent {
	} else if event == member.PostUnblockUserEvent {
	} else if event == member.PostSendFriendRequestEvent {
		r := data.(*member.SnsFriendRequest)
		log.Infof("got new friend request event, sending system and psn message")
		go c.SendMessageWithDataV2(r.ReceiverID, "info-new-friend-request", r.Receiver.Language, map[string]interface{}{"UserName": r.Sender.DisplayName}, map[string]string{
			"channel": "friends-request",
		}, true, true)
	} else if event == member.PostApprovedFriendRequestEvent {
		r := data.(*member.SnsFriendRequest)
		log.Infof("got new friend approved event, sending system and psn message")
		go c.SendMessageWithDataV2(r.SenderID, "info-friend-request-approved", r.Receiver.Language, map[string]interface{}{"UserName": r.Receiver.DisplayName}, map[string]string{
			"channel": "friends-request-sent",
		}, true, true)

	} else if event == member.PostRejectFriendRequestEvent {
		r := data.(*member.SnsFriendRequest)
		log.Infof("got new friend reject event, sending system and psn message")
		go c.SendMessageWithDataV2(r.SenderID, "info-friend-request-reject", r.Receiver.Language, map[string]interface{}{"UserName": r.Receiver.DisplayName}, map[string]string{
			"channel": "friends-request-sent",
		}, true, true)
	}
}
