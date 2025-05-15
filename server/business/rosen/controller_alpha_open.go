package rosen

import (
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"github.com/GoROSEN/rosen-apiserver/core/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/martian/log"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// setupOpenAlphaController 初始化控制器
func (c *Controller) setupOpenAlphaController(r *gin.RouterGroup) {

	r.GET("/alpha/plot/:id", c.plotDetail)
	r.GET("/alpha/ranking/blazers", c.getBlazerRankingBoard)
	r.GET("/alpha/ranking/producers", c.getProducerRankingBoard)
	r.GET("/alpha/ranking/m2e", c.getMove2EarnRankingBoard)
	r.GET("/alpha/ranking/game-challenges", c.getGameChallengerRankingBoard)
	r.GET("/alpha/ranking/cities", c.getCitiesRankingBoard)
	r.GET("/alpha/ranking/followers-increase", c.getFollowersIncreaseRankingBoard)
	r.GET("/alpha/custom-service/info", c.getCustomServiceInfo)
	r.GET("/alpha/ranking/followers", c.getFollowersRankingBoard)
	r.GET("/alpha/ranking/activities", c.getActivityRankingBoard)
	r.GET("/alpha/statistics", c.getAlphaStatistics)
}

func (c *Controller) plotDetail(ctx *gin.Context) {
	var plot Plot
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id < 1 {
		utils.SendFailureResponse(ctx, 500, "message.plot.invalid-plot-id")
		return
	}
	if err := c.Crud.GetPreloadModelByID(&plot, uint(id), []string{"Blazer", "Blazer.CurrentEquip", "Blazer.Member", "CoBlazer", "CoBlazer.Member", "Listing"}); err != nil {
		log.Errorf("cannot get plot: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.plot.plot-not-found")
		return
	}
	if err := c.Crud.Db.Model(&plot).Update("access_count", gorm.Expr("access_count + 1")).Error; err != nil {
		log.Errorf("cannot update access count: %v", err)
	}
}

func (c *Controller) getBlazerRankingBoard(ctx *gin.Context) {

	t := ctx.Query("type")
	var results []struct {
		ID          int    `json:"id"`
		Avatar      string `json:"avatar"` // 头像
		DisplayName string `json:"displayName"`
		Cnt         int    `json:"mintCount"`
		Description string `json:"-"`
		Equip       struct {
			ImageIndex uint   `json:"imgindex"`
			Logo       string `json:"logo"`
		} `gorm:"-" json:"equip"`
	}

	if t == "day" {
		if err := c.Crud.Db.Raw(`select t0.blazer_id as id, t1.avatar, t1.display_name,count(*) cnt, t3.description description
														 from rosen_mint_logs t0 
														 	 left join member_users t1 on t0.blazer_id = t1.id 
															 left join rosen_member_extras t2 on t1.id=t2.member_id 
															 left join rosen_assets t3 on t2.virtual_image_id = t3.id
														 where t0.blazer_id not in (6,3) 
														 			 and DATE_FORMAT( t0.created_at, '%Y%m%d' ) = DATE_FORMAT( UTC_DATE( ) , '%Y%m%d' ) 
														 group by blazer_id  
														 order by cnt desc 
														 limit 30`).Scan(&results).Error; err != nil {
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			log.Errorf("cannot get %v blazer ranking: %v", t, err)
			return
		}
	} else if t == "month" {
		if err := c.Crud.Db.Raw(`select t0.blazer_id as id, t1.avatar, t1.display_name,count(*) cnt, t3.description description
														 from rosen_mint_logs t0 
														   left join member_users t1 on t0.blazer_id = t1.id 
															 left join rosen_member_extras t2 on t1.id=t2.member_id 
															 left join rosen_assets t3 on t2.virtual_image_id = t3.id  
														 where t0.blazer_id not in (6,3) 
														       and DATE_FORMAT( t0.created_at, '%Y%m' ) = DATE_FORMAT( UTC_DATE( ) , '%Y%m' ) 
														 group by blazer_id
														 order by cnt desc
														 limit 30`).Scan(&results).Error; err != nil {
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			log.Errorf("cannot get %v blazer ranking: %v", t, err)
			return
		}
	} else {
		if err := c.Crud.Db.Raw(`select t0.blazer_id as id, t1.avatar, t1.display_name,count(*) cnt, t3.description description 
		 												 from rosen_mint_logs t0 
														   left join member_users t1 on t0.blazer_id = t1.id 
															 left join rosen_member_extras t2 on t1.id=t2.member_id 
															 left join rosen_assets t3 on t2.virtual_image_id = t3.id
														 where t0.blazer_id not in (6,3) 
														 group by blazer_id
														 order by cnt desc
														 limit 30`).Scan(&results).Error; err != nil {
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			log.Errorf("cannot get %v blazer ranking: %v", t, err)
			return
		}
	}
	for i := range results {
		if avatarURL, err := c.OssController.OssDownloadURLWithoutPrefix(results[i].Avatar); err != nil {
			log.Errorf("cannot get oss url: %v", err)
		} else {
			results[i].Avatar = avatarURL.String()
		}
		var sp SuitParamsVO
		if err := json.Unmarshal([]byte(results[i].Description), &sp); err != nil {
			log.Errorf("cannot unmarshal suit params vo: %v", err)
			imgindex, _ := strconv.Atoi(results[i].Description)
			results[i].Equip.ImageIndex = uint(imgindex)
		} else {
			results[i].Equip.ImageIndex = sp.ImageIndex
			results[i].Equip.Logo = sp.AvatarFrame
			// vo.Image = sp.Image
			// vo.SuitImage = sp.SuitImage
		}
	}

	utils.SendSuccessResponse(ctx, results)
}

func (c *Controller) getProducerRankingBoard(ctx *gin.Context) {

	t := ctx.Query("type")
	var results []struct {
		ID          int    `json:"id"`
		Avatar      string `json:"avatar"` // 头像
		DisplayName string `json:"displayName"`
		Cnt         int    `json:"mintCount"`
		Description string `json:"-"`
		Equip       struct {
			ImageIndex uint   `json:"imgindex"`
			Logo       string `json:"logo"`
		} `gorm:"-" json:"equip"`
	}

	if t == "day" {
		if err := c.Crud.Db.Raw(`select t0.producer_id as id,t1.avatar, t1.display_name,count(*) cnt, t3.description description 
														 from rosen_mint_logs t0 
														   left join member_users t1 on t0.producer_id = t1.id 
															 left join rosen_member_extras t2 on t1.id=t2.member_id 
															 left join rosen_assets t3 on t2.virtual_image_id = t3.id  
														 where t0.success=1 
														 			 and DATE_FORMAT( t0.created_at, '%Y%m%d' ) = DATE_FORMAT( UTC_DATE( ) , '%Y%m%d' ) 
														 group by producer_id  
														 order by cnt desc 
														 limit 30`).Scan(&results).Error; err != nil {
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			log.Errorf("cannot get %v producer ranking: %v", t, err)
			return
		}
	} else if t == "month" {
		if err := c.Crud.Db.Raw(`select t0.producer_id as id,t1.avatar, t1.display_name,count(*) cnt, t3.description description 
														 from rosen_mint_logs t0 
														   left join member_users t1 on t0.producer_id = t1.id 
															 left join rosen_member_extras t2 on t1.id=t2.member_id 
															 left join rosen_assets t3 on t2.virtual_image_id = t3.id  
														 where t0.success=1 
														 			 and DATE_FORMAT( t0.created_at, '%Y%m' ) = DATE_FORMAT( UTC_DATE( ) , '%Y%m' ) 
														 group by producer_id  
														 order by cnt desc
														 limit 30`).Scan(&results).Error; err != nil {
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			log.Errorf("cannot get %v producer ranking: %v", t, err)
			return
		}
	} else {
		if err := c.Crud.Db.Raw(`select t0.producer_id as id,t1.avatar, t1.display_name,count(*) cnt, t3.description description 
														 from rosen_mint_logs t0 
														   left join member_users t1 on t0.producer_id = t1.id
															 left join rosen_member_extras t2 on t1.id=t2.member_id 
															 left join rosen_assets t3 on t2.virtual_image_id = t3.id  
														 where t0.success=1 
														 group by producer_id 
														 order by cnt desc 
														 limit 30`).Scan(&results).Error; err != nil {
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			log.Errorf("cannot get %v producer ranking: %v", t, err)
			return
		}
	}
	for i := range results {
		if avatarURL, err := c.OssController.OssDownloadURLWithoutPrefix(results[i].Avatar); err != nil {
			log.Errorf("cannot get oss url: %v", err)
		} else {
			results[i].Avatar = avatarURL.String()
		}
		var sp SuitParamsVO
		if err := json.Unmarshal([]byte(results[i].Description), &sp); err != nil {
			log.Errorf("cannot unmarshal suit params vo: %v", err)
			imgindex, _ := strconv.Atoi(results[i].Description)
			results[i].Equip.ImageIndex = uint(imgindex)
		} else {
			results[i].Equip.ImageIndex = sp.ImageIndex
			results[i].Equip.Logo = sp.AvatarFrame
			// vo.Image = sp.Image
			// vo.SuitImage = sp.SuitImage
		}
	}
	utils.SendSuccessResponse(ctx, results)
}

func (c *Controller) getMove2EarnRankingBoard(ctx *gin.Context) {
	t := ctx.Query("type")
	var results []struct {
		ID          int    `json:"id"`
		Avatar      string `json:"avatar"` // 头像
		DisplayName string `json:"displayName"`
		Cnt         int    `json:"earned"`
		Description string `json:"-"`
		Equip       struct {
			ImageIndex uint   `json:"imgindex"`
			Logo       string `json:"logo"`
		} `gorm:"-" json:"equip"`
	}

	if t == "day" {
		if err := c.Crud.Db.Raw(`select t0.member_id as id,t1.avatar, t1.display_name,floor(sum(earned)/100) cnt, t3.description description 
															from rosen_mte_sessions t0 
																left join member_users t1 on t0.member_id = t1.id 
															 	left join rosen_member_extras t2 on t1.id=t2.member_id 
																left join rosen_assets t3 on t2.virtual_image_id = t3.id  
															where DATE_FORMAT( t0.created_at, '%Y%m%d' ) = DATE_FORMAT( UTC_DATE( ) , '%Y%m%d' ) 
															group by t0.member_id 
															order by cnt desc 
															limit 30`).Scan(&results).Error; err != nil {
			log.Errorf("cannot get %v m2e ranking: %v", t, err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	} else if t == "month" {
		if err := c.Crud.Db.Raw(`select t0.member_id as id,t1.avatar, t1.display_name,floor(sum(earned)/100) cnt, t3.description description 
															from rosen_mte_sessions t0 
																left join member_users t1 on t0.member_id = t1.id 
															 	left join rosen_member_extras t2 on t1.id=t2.member_id 
																left join rosen_assets t3 on t2.virtual_image_id = t3.id  
															where DATE_FORMAT( t0.created_at, '%Y%m' ) = DATE_FORMAT( UTC_DATE( ) , '%Y%m' ) 
															group by t0.member_id  
															order by cnt desc 
															limit 30`).Scan(&results).Error; err != nil {
			log.Errorf("cannot get %v m2e ranking: %v", t, err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	} else {
		if err := c.Crud.Db.Raw(`select t0.member_id as id,t1.avatar, t1.display_name,floor(sum(earned)/100) cnt, t3.description description 
															from rosen_mte_sessions t0 
																left join member_users t1 on t0.member_id = t1.id 
															 	left join rosen_member_extras t2 on t1.id=t2.member_id 
																left join rosen_assets t3 on t2.virtual_image_id = t3.id  
															group by t0.member_id  
															order by cnt desc 
															limit 30`).Scan(&results).Error; err != nil {
			log.Errorf("cannot get %v m2e ranking: %v", t, err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	}
	for i := range results {
		if avatarURL, err := c.OssController.OssDownloadURLWithoutPrefix(results[i].Avatar); err != nil {
			log.Errorf("cannot get oss url: %v", err)
		} else {
			results[i].Avatar = avatarURL.String()
		}
		var sp SuitParamsVO
		if err := json.Unmarshal([]byte(results[i].Description), &sp); err != nil {
			log.Errorf("cannot unmarshal suit params vo: %v", err)
			imgindex, _ := strconv.Atoi(results[i].Description)
			results[i].Equip.ImageIndex = uint(imgindex)
		} else {
			results[i].Equip.ImageIndex = sp.ImageIndex
			results[i].Equip.Logo = sp.AvatarFrame
			// vo.Image = sp.Image
			// vo.SuitImage = sp.SuitImage
		}
	}
	utils.SendSuccessResponse(ctx, results)
}

func (c *Controller) getGameChallengerRankingBoard(ctx *gin.Context) {

	t := ctx.Query("type")
	var results []struct {
		ID             int    `json:"id"`
		Avatar         string `json:"avatar"` // 头像
		DisplayName    string `json:"displayName"`
		ChallengeTimes int    `json:"challengeTimes"`
		Description    string `json:"-"`
		Equip          struct {
			ImageIndex uint   `json:"imgindex"`
			Logo       string `json:"logo"`
		} `gorm:"-" json:"equip"`
	}

	if t == "day" {
		if err := c.Crud.Db.Raw(`select t0.host_id as id,t1.avatar, t1.display_name,count(1) as challenge_times, t3.description description 
															from rosen_game_sessions t0 
																left join member_users t1 on t0.host_id = t1.id 
															 	left join rosen_member_extras t2 on t1.id=t2.member_id 
																left join rosen_assets t3 on t2.virtual_image_id = t3.id  
															where t0.status=2 and DATE_FORMAT( t0.created_at, '%Y%m%d' ) = DATE_FORMAT( UTC_DATE( ) , '%Y%m%d' ) 
															group by host_id  
															order by challenge_times desc 
															limit 30`).Scan(&results).Error; err != nil {
			log.Errorf("cannot get %v game challenger ranking: %v", t, err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	} else if t == "month" {
		if err := c.Crud.Db.Raw(`select t0.host_id as id,t1.avatar, t1.display_name,count(1) as challenge_times, t3.description description 
															from rosen_game_sessions t0 
																left join member_users t1 on t0.host_id = t1.id 
															 	left join rosen_member_extras t2 on t1.id=t2.member_id 
																left join rosen_assets t3 on t2.virtual_image_id = t3.id  
															where t0.status=2 and DATE_FORMAT( t0.created_at, '%Y%m' ) = DATE_FORMAT( UTC_DATE( ) , '%Y%m' ) 
															group by host_id  
															order by challenge_times 
															desc limit 30`).Scan(&results).Error; err != nil {
			log.Errorf("cannot get %v challenger ranking: %v", t, err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	} else {
		if err := c.Crud.Db.Raw(`select t0.host_id as id,t1.avatar, t1.display_name,count(1) as challenge_times, t3.description description 
															from rosen_game_sessions t0 
																left join member_users t1 on t0.host_id = t1.id 
															 	left join rosen_member_extras t2 on t1.id=t2.member_id 
																left join rosen_assets t3 on t2.virtual_image_id = t3.id  
															where t0.status=2 
															group by host_id 
															order by challenge_times 
															desc limit 30`).Scan(&results).Error; err != nil {
			log.Errorf("cannot get %v challenger ranking: %v", t, err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	}
	log.Infof("%v", results)
	for i := range results {
		if avatarURL, err := c.OssController.OssDownloadURLWithoutPrefix(results[i].Avatar); err != nil {
			log.Errorf("cannot get oss url: %v", err)
		} else {
			results[i].Avatar = avatarURL.String()
		}
		var sp SuitParamsVO
		if err := json.Unmarshal([]byte(results[i].Description), &sp); err != nil {
			log.Errorf("cannot unmarshal suit params vo: %v", err)
			imgindex, _ := strconv.Atoi(results[i].Description)
			results[i].Equip.ImageIndex = uint(imgindex)
		} else {
			results[i].Equip.ImageIndex = sp.ImageIndex
			results[i].Equip.Logo = sp.AvatarFrame
			// vo.Image = sp.Image
			// vo.SuitImage = sp.SuitImage
		}
	}
	utils.SendSuccessResponse(ctx, results)
}

func (c *Controller) getCitiesRankingBoard(ctx *gin.Context) {

	type _ResultItem struct {
		ID            int       `json:"id"`
		Avatar        string    `json:"avatar"` // 头像
		DisplayName   string    `json:"displayName"`
		City          string    `json:"city"`
		PlotName      string    `json:"plot"`
		PlotLogo      string    `json:"plotLogo"`
		MintCount     int64     `json:"mintCount"`
		Rank          int       `json:"rank"`
		CreatedAt     time.Time `json:"-"`
		ExpiredAt     time.Time `json:"-"`
		PlotID        uint      `json:"plotId"`
		BlazerID      uint      `json:"blazerId"`
		PlotLatitude  float64   `json:"plotLat"`
		PlotLongitude float64   `json:"plotLng"`
		Description   string    `json:"-"`
		Equip         struct {
			ImageIndex uint   `json:"imgindex"`
			Logo       string `json:"logo"`
		} `gorm:"-" json:"equip"`
	}
	t := ctx.Query("type")
	var results []_ResultItem

	if err := c.Crud.Db.Raw(`select t0.created_at, t0.expired_at, t0.id, t3.avatar, t3.id as blazer_id, t1.id as plot_id, t1.longitude as plot_longitude, t1.latitude as plot_latitude, t3.display_name, t0.city,t1.name as plot_name, t1.logo as plot_logo, t4.description description 
														from rosen_cities_ranks t0 
															left join rosen_plots t1 on t0.plot_id = t1.id
															left join rosen_member_extras t2 on t0.blazer_id = t2.member_id
															left join member_users t3 on t0.blazer_id = t3.id
															left join rosen_assets t4 on t2.virtual_image_id = t4.id  
														where t0.deleted_at is null
														order by mint_count desc`).Scan(&results).Error; err != nil {
		log.Errorf("cannot get %v m2e ranking: %v", t, err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}

	// 挨个获取自统计时间起的mint数量
	if t == "day" {
		for i := range results {
			r := &results[i]
			if err := c.Crud.Db.Model(&MintLog{}).Where("plot_id = ? and success = 1 and created_at >= ? and created_at <= ? and DATE_FORMAT(created_at, '%Y%m%d' ) = DATE_FORMAT( UTC_DATE( ) , '%Y%m%d' )", r.PlotID, r.CreatedAt, r.ExpiredAt).Count(&r.MintCount).Error; err != nil {
				log.Errorf("cannot get mint count for plot %v : %v", r.PlotID, err)
				r.MintCount = 0
			}
		}
	} else if t == "month" {
		for i := range results {
			r := &results[i]
			if err := c.Crud.Db.Model(&MintLog{}).Where("plot_id = ? and success = 1 and created_at >= ? and created_at <= ? and DATE_FORMAT(created_at, '%Y%m' ) = DATE_FORMAT( UTC_DATE( ) , '%Y%m' )", r.PlotID, r.CreatedAt, r.ExpiredAt).Count(&r.MintCount).Error; err != nil {
				log.Errorf("cannot get mint count for plot %v : %v", r.PlotID, err)
				r.MintCount = 0
			}
		}
	} else {
		for i := range results {
			r := &results[i]
			if err := c.Crud.Db.Model(&MintLog{}).Where("plot_id = ? and success = 1 and created_at >= ? and created_at <= ?", r.PlotID, r.CreatedAt, r.ExpiredAt).Count(&r.MintCount).Error; err != nil {
				log.Errorf("cannot get mint count for plot %v : %v", r.PlotID, err)
				r.MintCount = 0
			}
		}
	}

	// 排个序
	sort.Slice(results, func(i, j int) bool {
		return results[i].MintCount > results[j].MintCount
	})

	for i := range results {
		results[i].Rank = i + 1
		if avatarURL, err := c.OssController.OssDownloadURLWithoutPrefix(results[i].Avatar); err != nil {
			log.Errorf("cannot get oss url: %v", err)
		} else {
			results[i].Avatar = avatarURL.String()
		}
		if logoURL, err := c.OssController.OssDownloadURL(results[i].PlotLogo); err != nil {
			log.Errorf("cannot get oss url: %v", err)
		} else {
			results[i].PlotLogo = logoURL.String()
		}
		var sp SuitParamsVO
		if err := json.Unmarshal([]byte(results[i].Description), &sp); err != nil {
			log.Errorf("cannot unmarshal suit params vo: %v", err)
			imgindex, _ := strconv.Atoi(results[i].Description)
			results[i].Equip.ImageIndex = uint(imgindex)
		} else {
			results[i].Equip.ImageIndex = sp.ImageIndex
			results[i].Equip.Logo = sp.AvatarFrame
			// vo.Image = sp.Image
			// vo.SuitImage = sp.SuitImage
		}
	}
	utils.SendSuccessResponse(ctx, results)
}

func (c *Controller) getCustomServiceInfo(ctx *gin.Context) {
	var extra MemberExtra
	id := c.service.sysconfig.CustomServiceMemberId
	if id < 1 {
		utils.SendFailureResponse(ctx, 400, "message.common.system-error")
		return
	}
	if err := c.Crud.FindPreloadModelWhere(&extra, []string{"CurrentEquip", "Member", "Privileges"}, "member_id = ?", id); err != nil {
		log.Errorf("cannot get member by token: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.member.member-not-found")
		return
	}

	var vo struct {
		MemberFullVO
	}
	tvo := c.composeRosenMemberFullVO(&extra, nil, nil, nil, nil, nil)
	tvo.UserName = "***"
	if len(tvo.Email) > 5 {
		tvo.Email = tvo.Email[0:1] + "***" + tvo.Email[len(tvo.Email)-4:len(tvo.Email)]
	} else {
		tvo.Email = "***"
	}
	copier.Copy(&vo, tvo)

	utils.SendSuccessResponse(ctx, vo)
}

func (c *Controller) getFollowersRankingBoard(ctx *gin.Context) {

	type _ResultItem struct {
		ID             int    `json:"id"`
		Rank           int    `json:"rank"`
		Avatar         string `json:"avatar"` // 头像
		DisplayName    string `json:"displayName"`
		FollowersCount int64  `json:"followersCount"`
		Description    string `json:"-"`
		Equip          struct {
			ImageIndex uint   `json:"imgindex"`
			Logo       string `json:"logo"`
		} `gorm:"-" json:"equip"`
	}
	results := []_ResultItem{}

	if err := c.Crud.Db.Raw(`select t0.member_id as id, count(t0.follower_id) as followers_count, t1.display_name, t1.avatar, t3.description description 
														from member_sns_followers t0 
															left join member_users t1 on t0.member_id = t1.id 
															left join rosen_member_extras t2 on t1.id=t2.member_id 
															left join rosen_assets t3 on t2.virtual_image_id = t3.id  
														where t0.member_id != ? 
														group by t0.member_id 
														order by followers_count desc 
														limit 30;`, c.service.sysconfig.CustomServiceMemberId).Scan(&results).Error; err != nil {
		log.Errorf("cannot get followers ranking: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}

	for i := range results {
		results[i].Rank = i + 1
		if avatarURL, err := c.OssController.OssDownloadURLWithoutPrefix(results[i].Avatar); err != nil {
			log.Errorf("cannot get oss url: %v", err)
		} else {
			results[i].Avatar = avatarURL.String()
		}
		var sp SuitParamsVO
		if err := json.Unmarshal([]byte(results[i].Description), &sp); err != nil {
			log.Errorf("cannot unmarshal suit params vo: %v", err)
			imgindex, _ := strconv.Atoi(results[i].Description)
			results[i].Equip.ImageIndex = uint(imgindex)
		} else {
			results[i].Equip.ImageIndex = sp.ImageIndex
			results[i].Equip.Logo = sp.AvatarFrame
			// vo.Image = sp.Image
			// vo.SuitImage = sp.SuitImage
		}
	}
	utils.SendSuccessResponse(ctx, results)
}

func (c *Controller) getFollowersIncreaseRankingBoard(ctx *gin.Context) {

	type _ResultItem struct {
		ID            int    `json:"id"`
		Rank          int    `json:"rank"`
		Avatar        string `json:"avatar"` // 头像
		DisplayName   string `json:"displayName"`
		IncreaseCount int64  `json:"increaseCount"`
		Description   string `json:"-"`
		Equip         struct {
			ImageIndex uint   `json:"imgindex"`
			Logo       string `json:"logo"`
		} `gorm:"-" json:"equip"`
	}
	results := []_ResultItem{}

	boardType := ctx.Query("type")
	dueTo := time.Now().Format("2006-01-02")
	var startsFrom string
	if boardType == "daily" {
		startsFrom = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	} else {
		startsFrom = time.Now().Format("2006-01-") + "01"
	}
	log.Debugf("starts from %v to %v", startsFrom, dueTo)

	if err := c.Crud.Db.Raw(`SELECT t0.member_id as id, t1.friends_count - t0.friends_count AS increase_count, t2.display_name, t2.avatar, t4.description description
														FROM rosen_member_friends_trend_snapshot t0
														LEFT JOIN
														(
															SELECT member_id,
															friends_count	
															FROM rosen_member_friends_trend_snapshot	
															WHERE DATE_FORMAT(created_at, '%Y-%m-%d') = ?
														)
														t1 ON t0.member_id = t1.member_id
														LEFT JOIN member_users t2 ON t0.member_id = t2.id
														LEFT JOIN rosen_member_extras t3 ON t2.id = t3.member_id
														LEFT JOIN rosen_assets t4 ON t3.virtual_image_id = t4.id
														WHERE t0.member_id != ? and DATE_FORMAT(t0.created_at, '%Y-%m-%d') = ?
														ORDER BY increase_count DESC LIMIT 30`, dueTo, c.service.sysconfig.CustomServiceMemberId, startsFrom).Scan(&results).Error; err != nil {
		log.Errorf("cannot get followers increasements: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}

	log.Debugf("got %v results", len(results))

	for i := range results {
		results[i].Rank = i + 1
		if avatarURL, err := c.OssController.OssDownloadURLWithoutPrefix(results[i].Avatar); err != nil {
			log.Errorf("cannot get oss url: %v", err)
		} else {
			results[i].Avatar = avatarURL.String()
		}
		var sp SuitParamsVO
		if err := json.Unmarshal([]byte(results[i].Description), &sp); err != nil {
			log.Errorf("cannot unmarshal suit params vo: %v", err)
			imgindex, _ := strconv.Atoi(results[i].Description)
			results[i].Equip.ImageIndex = uint(imgindex)
		} else {
			results[i].Equip.ImageIndex = sp.ImageIndex
			results[i].Equip.Logo = sp.AvatarFrame
			// vo.Image = sp.Image
			// vo.SuitImage = sp.SuitImage
		}
	}
	utils.SendSuccessResponse(ctx, results)
}

func (c *Controller) getAlphaStatistics(ctx *gin.Context) {
	/*
		1. 用户来自xx个国家
		2. 使用xx种语言
		3. 有多少用户在rosen上赚到钱（所有吧？因为现在每个人都有0.5u）
		4. 有多少用户在rosen上拥有数字资产（好像也是所有人？或者有皮肤或mint过nft的人）
		5. 在Rosen星球产生的资产总和（地块的价格总和+二手地块的交易量+小游戏的交易总量+mint nft的交易量+每个用户0.5usdt的token）
		6. NFT mint的总数量
	*/

	var result struct {
		Users struct {
			CountriesCount int64 `json:"countriesCount"`
			LanguagesCount int64 `json:"languagesCount"`
			EarnedCount    int64 `json:"earnedCount"`
			NftOwnerCount  int64 `json:"nftOwners"`
		} `json:"users"`
		Assets struct {
			ProducedCount  int64 `json:"producedCount"`
			MintedNftCount int64 `json:"mintedNftCount"`
		} `json:"assets"`
	}
	if err := c.Crud.Db.Raw(`select SUBSTRING_INDEX(language,'-',-1) from member_users group by SUBSTRING_INDEX(language,'-',-1)`).Count(&result.Users.CountriesCount).Error; err != nil {
		log.Errorf("cannot get User Countries Count: %v", err)
	}
	if err := c.Crud.Db.Raw(`select SUBSTRING_INDEX(language,'-',1) from member_users group by SUBSTRING_INDEX(language,'-',1)`).Count(&result.Users.LanguagesCount).Error; err != nil {
		log.Errorf("cannot get User Languages Count: %v", err)
	}
	if err := c.Crud.Db.Model(&MemberExtra{}).Count(&result.Users.EarnedCount).Error; err != nil {
		log.Errorf("cannot get Member Extra Count: %v", err)
	}
	if err := c.Crud.Db.Raw(`select owner_id,count(1) from rosen_assets where nft_address != '' group by owner_id`).Count(&result.Users.NftOwnerCount).Error; err != nil {
		log.Errorf("cannot get nft owner Count: %v", err)
	}
	if err := c.Crud.Db.Model(&Asset{}).Where("nft_address != ''").Count(&result.Assets.MintedNftCount).Error; err != nil {
		log.Errorf("cannot get Asset Count: %v", err)
	}

	utils.SendSuccessResponse(ctx, result)
}

func (c *Controller) getHallChatActives(ctx *gin.Context) {
	var results []struct {
		FromUserId uint
		Cnt        int
	}
	const lastMessageCount = 100
	if err := c.Crud.Db.Raw(`SELECT from_user_id, count(1) as cnt
														FROM 
														(select * from rosen_chat_histories order by created_at desc limit ?) as hist
														WHERE deleted_at IS NULL
														AND channel = 'chat'
														AND action = 'hallchat'
														GROUP BY from_user_id
														order by cnt desc`, lastMessageCount).Scan(&results).Error; err != nil {
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		log.Errorf("cannot get hall chat actives: %v", err)
		return
	}
	memberIds := make([]uint, len(results))
	for i := range results {
		memberIds[i] = results[i].FromUserId
	}
	var members []MemberExtra
	if err := c.Crud.List(&members, []string{"Member", "CurrentEquip"}, "", "member_id in ?", memberIds); err != nil {
		log.Errorf("cannot find active members")
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}

	vo := c.memberExtraList2MemberWithEquipVOList(&members).([]*MemberWithEquipVO)

	utils.SendSuccessResponse(ctx, vo)
}

func (c *Controller) getActivityRankingBoard(ctx *gin.Context) {

	type _ResultItem struct {
		ID            int    `json:"id"`
		Rank          int    `json:"rank"`
		Avatar        string `json:"avatar"` // 头像
		DisplayName   string `json:"displayName"`
		ActivityScore int64  `json:"activityScore"`
		Description   string `json:"-"`
		Equip         struct {
			ImageIndex uint   `json:"imgindex"`
			Logo       string `json:"logo"`
		} `gorm:"-" json:"equip"`
	}
	results := []_ResultItem{}

	boardType := ctx.Query("type")
	dueTo := time.Now().Format("2006-01-02")
	var startsFrom string
	if boardType == "daily" {
		startsFrom = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	} else {
		startsFrom = time.Now().Format("2006-01-") + "01"
	}

	if err := c.Crud.Db.Raw(`select t0.member_id as id, sum(score) as score, t1.display_name, t1.avatar, t3.description description 
														from rosen_activity_trends t0 
															left join member_users t1 on t0.member_id = t1.id 
															left join rosen_member_extras t2 on t1.id=t2.member_id 
															left join rosen_assets t3 on t2.virtual_image_id = t3.id  
														where t0.member_id != ? and t0.created_at >= ? and t0.created_at < ?
														group by t0.member_id 
														order by score desc 
														limit 30`, c.service.sysconfig.CustomServiceMemberId, startsFrom, dueTo).Scan(&results).Error; err != nil {
		log.Errorf("cannot get followers ranking: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}

	for i := range results {
		results[i].Rank = i + 1
		if avatarURL, err := c.OssController.OssDownloadURLWithoutPrefix(results[i].Avatar); err != nil {
			log.Errorf("cannot get oss url: %v", err)
		} else {
			results[i].Avatar = avatarURL.String()
		}
		var sp SuitParamsVO
		if err := json.Unmarshal([]byte(results[i].Description), &sp); err != nil {
			log.Errorf("cannot unmarshal suit params vo: %v", err)
			imgindex, _ := strconv.Atoi(results[i].Description)
			results[i].Equip.ImageIndex = uint(imgindex)
		} else {
			results[i].Equip.ImageIndex = sp.ImageIndex
			results[i].Equip.Logo = sp.AvatarFrame
			// vo.Image = sp.Image
			// vo.SuitImage = sp.SuitImage
		}
	}
	utils.SendSuccessResponse(ctx, results)
}
