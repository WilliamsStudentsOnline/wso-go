package goodrich

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/email"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/go-mail/mail"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

const DateFormat = "2006-01-02"
const TimeSlotFormat = "%02d:%02d"

var BannedTimeSlot = NewTimeSlotInt(10, 10)

type Controller struct {
	services.BaseController
	// Put a model here, like:
	orderModel *models.GoodrichOrderModel
	menuModel  *models.GoodrichMenuItemModel
	userModel  *models.UserModel
	cfg        *config.Config
	email      *email.GoodrichMailer
}

// Construct a new user controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
		orderModel:     models.NewGoodrichOrderModel(db, log),
		menuModel:      models.NewGoodrichMenuItemModel(db, log),
		userModel:      models.NewUserModel(db, log),
		cfg:            cfg,
		email:          email.NewGoodrichMailer(cfg, log.Named("email")),
	}
}

// ListTimeSlots godoc
// @Summary List time slots
// @Description lists all time slots today
// @ID goodrich-list-time-slots
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Success 200 {array} goodrich.TimeSlot
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /goodrich/timeslots [get]
func (t *Controller) ListTimeSlots(c *gin.Context) {
	date := time.Now().Format(DateFormat)

	// return nothing if this date is closed
	foundValidDate := false
	for _, date := range t.cfg.GoodrichOpenDays {
		if time.Now().Format(DateFormat) == date {
			foundValidDate = true
			break
		}
	}
	if !foundValidDate {
		t.RespondOK(c, []TimeSlot{})
		return
	}

	tsMap, err := t.generateSlotAvailabilityMap(date)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	slots := t.generateTimeSlotsAfter(time.Now())
	for _, slot := range slots {
		slot.ClosedSpots = uint(tsMap[slot.String()])
		slot.OpenSpots = uint(t.cfg.GoodrichSlotSpotSize) - slot.ClosedSpots
		slot.Formatted = slot.String()
	}

	t.RespondOK(c, slots)
}

func (t *Controller) generateSlotAvailabilityMap(date string) (tsMap map[string]int, err error) {
	var orders []*models.GoodrichOrder
	err = t.orderModel.GetOrdersByDate(&orders, date)
	if err != nil {
		return
	}

	tsMap = make(map[string]int)

	for _, order := range orders {
		val, ok := tsMap[order.TimeSlot]
		if !ok {
			tsMap[order.TimeSlot] = 1
		} else {
			tsMap[order.TimeSlot] = val + 1
		}
	}

	return
}

type TimeSlot struct {
	Hour        uint   `json:"hour"`
	Minute      uint   `json:"minute"`
	Formatted   string `json:"formatted"`
	ClosedSpots uint   `json:"closedSpots"`
	OpenSpots   uint   `json:"openSpots"`
}

func NewTimeSlot(hour, minute uint) *TimeSlot {
	return &TimeSlot{
		Hour:   hour,
		Minute: minute,
	}
}

func NewTimeSlotInt(hour, minute int) *TimeSlot {
	return NewTimeSlot(uint(hour), uint(minute))
}

func (g TimeSlot) Clone() *TimeSlot {
	return NewTimeSlot(g.Hour, g.Minute)
}

func (g *TimeSlot) increment() {
	g.Minute += 10
	if g.Minute >= 60 {
		g.Hour += 1
		g.Minute = g.Minute - 60
	}
}

func (g *TimeSlot) Equal(o *TimeSlot) bool {
	return g.Hour == o.Hour && g.Minute == o.Minute
}

func (g *TimeSlot) After(o *TimeSlot) bool {
	return g.Hour > o.Hour || (g.Hour == o.Hour && g.Minute > o.Minute)
}

func (g *TimeSlot) String() string {
	return fmt.Sprintf(TimeSlotFormat, g.Hour, g.Minute)
}

func (t *Controller) generateAllDailyTimeSlots() (slots []*TimeSlot) {
	end := NewTimeSlotInt(t.cfg.GoodrichMustParseTime(t.cfg.GoodrichClose))

	idx := NewTimeSlotInt(t.cfg.GoodrichMustParseTime(t.cfg.GoodrichOpen))
	idx.increment()

	for !idx.After(end) {
		if idx.Equal(BannedTimeSlot) {
			continue
		}
		slots = append(slots, idx.Clone())
		idx.increment()
	}

	return
}

func (t *Controller) generateTimeSlotsAfter(tm time.Time) (slots []*TimeSlot) {
	end := NewTimeSlotInt(t.cfg.GoodrichMustParseTime(t.cfg.GoodrichClose))
	tSlot := NewTimeSlotInt(tm.Hour(), tm.Minute())
	tSlot.increment() // Make sure to not allow 10 min after rn for a slot
	idx := NewTimeSlotInt(t.cfg.GoodrichMustParseTime(t.cfg.GoodrichOpen))
	idx.increment() // Make sure the opening slot is closed

	for !idx.After(end) {
		if idx.Equal(BannedTimeSlot) {
			idx.increment()
			continue
		}
		if idx.After(tSlot) || idx.Equal(tSlot) {
			slots = append(slots, idx.Clone())
		}
		idx.increment()
	}

	return
}

func (t *Controller) generateNotifEmail(order models.GoodrichOrder, userID uint) error {
	user := models.User{}
	err := t.userModel.GetUserByID(userID, &user)
	if err != nil {
		return err
	}

	m := mail.NewMessage()
	m.SetAddressHeader("From", "goodrich@wso.williams.edu", "WSO Goodrich")

	addrTo := user.WilliamsEmail
	if addrTo == "" {
		addrTo = user.UnixID + "@williams.edu"
	}
	m.SetAddressHeader("To", addrTo, user.Name)

	m.SetHeader("Subject", fmt.Sprintf("WSO Goodrich - Order %d Confirmation", order.ID))

	tmpl, err := template.New("email").Parse(notifEmailTemplate)
	if err != nil {
		return err
	}

	orderDate, err := time.ParseInLocation(DateFormat, order.Date, time.Local)
	if err != nil {
		return err
	}

	orderTime, err := time.ParseInLocation("15:04", order.TimeSlot, time.Local)
	if err != nil {
		return err
	}

	swipeOrder := order.PaymentMethod == models.GoodrichPaymentMethodSwipe || order.PaymentMethod == models.GoodrichPaymentMethodSwipePlusCreditCard || order.PaymentMethod == models.GoodrichPaymentMethodSwipePlusCash
	totalPriceAdj := order.TotalPrice
	if swipeOrder {
		totalPriceAdj = math.Max(0, order.TotalPrice-5)
	}

	paymentStr := "Unknown"
	switch order.PaymentMethod {
	case models.GoodrichPaymentMethodSwipe:
		paymentStr = "Swipe"
	case models.GoodrichPaymentMethodCreditCard:
		paymentStr = "Credit Card"
	case models.GoodrichPaymentMethodCash:
		paymentStr = "Cash"
	case models.GoodrichPaymentMethodSwipePlusCash:
		paymentStr = "Swipe + Cash"
	case models.GoodrichPaymentMethodSwipePlusCreditCard:
		paymentStr = "Swipe + Credit Card"
	}

	data := struct {
		User               models.User
		Order              models.GoodrichOrder
		OrderDate          string
		OrderTime          string
		TimeNow            string
		SwipeOrder         bool
		TotalPriceAdjusted float64
		PaymentString      string
	}{
		User:               user,
		Order:              order,
		OrderDate:          orderDate.Format("Monday, 1/2/2006"),
		OrderTime:          orderTime.Format("3:04PM"),
		TimeNow:            time.Now().Format("Jan _2 3:04 pm"),
		SwipeOrder:         swipeOrder,
		TotalPriceAdjusted: totalPriceAdj,
		PaymentString:      paymentStr,
	}

	var tmplBuf bytes.Buffer
	if err := tmpl.Execute(&tmplBuf, data); err != nil {
		return err
	}

	m.SetBody("text/html", tmplBuf.String())

	t.email.SendMail(m)

	return nil
}

const notifEmailTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
  <head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta name="x-apple-disable-message-reformatting" />
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <meta name="color-scheme" content="light dark" />
    <meta name="supported-color-schemes" content="light dark" />
    <title></title>
    <style type="text/css" rel="stylesheet" media="all">
    /* Base ------------------------------ */
    
    @import url("https://fonts.googleapis.com/css?family=Nunito+Sans:400,700&amp;display=swap");
    body {
      width: 100% !important;
      height: 100%;
      margin: 0;
      -webkit-text-size-adjust: none;
    }
    
    a {
      color: #3869D4;
    }
    
    a img {
      border: none;
    }
    
    td {
      word-break: break-word;
    }
    
    .preheader {
      display: none !important;
      visibility: hidden;
      mso-hide: all;
      font-size: 1px;
      line-height: 1px;
      max-height: 0;
      max-width: 0;
      opacity: 0;
      overflow: hidden;
    }
    /* Type ------------------------------ */
    
    body,
    td,
    th {
      font-family: "Nunito Sans", Helvetica, Arial, sans-serif;
    }
    
    h1 {
      margin-top: 0;
      color: #333333;
      font-size: 22px;
      font-weight: bold;
      text-align: left;
    }
    
    h2 {
      margin-top: 0;
      color: #333333;
      font-size: 16px;
      font-weight: bold;
      text-align: left;
    }
    
    h3 {
      margin-top: 0;
      color: #333333;
      font-size: 14px;
      font-weight: bold;
      text-align: left;
    }
    
    td,
    th {
      font-size: 16px;
    }
    
    p,
    ul,
    ol,
    blockquote {
      margin: .4em 0 1.1875em;
      font-size: 16px;
      line-height: 1.625;
    }
    
    p.sub {
      font-size: 13px;
    }
    /* Utilities ------------------------------ */
    
    .align-right {
      text-align: right;
    }
    
    .align-left {
      text-align: left;
    }
    
    .align-center {
      text-align: center;
    }
    /* Buttons ------------------------------ */
    
    .button {
      background-color: #3869D4;
      border-top: 10px solid #3869D4;
      border-right: 18px solid #3869D4;
      border-bottom: 10px solid #3869D4;
      border-left: 18px solid #3869D4;
      display: inline-block;
      color: #FFF;
      text-decoration: none;
      border-radius: 3px;
      box-shadow: 0 2px 3px rgba(0, 0, 0, 0.16);
      -webkit-text-size-adjust: none;
      box-sizing: border-box;
    }
    
    .button--green {
      background-color: #22BC66;
      border-top: 10px solid #22BC66;
      border-right: 18px solid #22BC66;
      border-bottom: 10px solid #22BC66;
      border-left: 18px solid #22BC66;
    }
    
    .button--red {
      background-color: #FF6136;
      border-top: 10px solid #FF6136;
      border-right: 18px solid #FF6136;
      border-bottom: 10px solid #FF6136;
      border-left: 18px solid #FF6136;
    }
    
    @media only screen and (max-width: 500px) {
      .button {
        width: 100% !important;
        text-align: center !important;
      }
    }
    /* Attribute list ------------------------------ */
    
    .attributes {
      margin: 0 0 21px;
    }
    
    .attributes_content {
      background-color: #F4F4F7;
      padding: 16px;
    }
    
    .attributes_item {
      padding: 0;
    }
    /* Related Items ------------------------------ */
    
    .related {
      width: 100%;
      margin: 0;
      padding: 25px 0 0 0;
      -premailer-width: 100%;
      -premailer-cellpadding: 0;
      -premailer-cellspacing: 0;
    }
    
    .related_item {
      padding: 10px 0;
      color: #CBCCCF;
      font-size: 15px;
      line-height: 18px;
    }
    
    .related_item-title {
      display: block;
      margin: .5em 0 0;
    }
    
    .related_item-thumb {
      display: block;
      padding-bottom: 10px;
    }
    
    .related_heading {
      border-top: 1px solid #CBCCCF;
      text-align: center;
      padding: 25px 0 10px;
    }
    /* Discount Code ------------------------------ */
    
    .discount {
      width: 100%;
      margin: 0;
      padding: 24px;
      -premailer-width: 100%;
      -premailer-cellpadding: 0;
      -premailer-cellspacing: 0;
      background-color: #F4F4F7;
      border: 2px dashed #CBCCCF;
    }
    
    .discount_heading {
      text-align: center;
    }
    
    .discount_body {
      text-align: center;
      font-size: 15px;
    }
    /* Social Icons ------------------------------ */
    
    .social {
      width: auto;
    }
    
    .social td {
      padding: 0;
      width: auto;
    }
    
    .social_icon {
      height: 20px;
      margin: 0 8px 10px 8px;
      padding: 0;
    }
    /* Data table ------------------------------ */
    
    .purchase {
      width: 100%;
      margin: 0;
      padding: 35px 0;
      -premailer-width: 100%;
      -premailer-cellpadding: 0;
      -premailer-cellspacing: 0;
    }
    
    .purchase_content {
      width: 100%;
      margin: 0;
      padding: 25px 0 0 0;
      -premailer-width: 100%;
      -premailer-cellpadding: 0;
      -premailer-cellspacing: 0;
    }
    
    .purchase_item {
      padding: 10px 0;
      color: #51545E;
      font-size: 15px;
      line-height: 18px;
    }
    
    .purchase_heading {
      padding-bottom: 8px;
      border-bottom: 1px solid #EAEAEC;
    }
    
    .purchase_heading p {
      margin: 0;
      color: #85878E;
      font-size: 12px;
    }
    
    .purchase_footer {
      padding-top: 15px;
      border-top: 1px solid #EAEAEC;
    }
    
    .purchase_total {
      margin: 0;
      text-align: right;
      font-weight: bold;
      color: #333333;
    }
    
    .purchase_total--label {
      padding: 0 15px 0 0;
    }
    
    body {
      background-color: #FFF;
      color: #333;
    }
    
    p {
      color: #333;
    }
    
    .email-wrapper {
      width: 100%;
      margin: 0;
      padding: 0;
      -premailer-width: 100%;
      -premailer-cellpadding: 0;
      -premailer-cellspacing: 0;
    }
    
    .email-content {
      width: 100%;
      margin: 0;
      padding: 0;
      -premailer-width: 100%;
      -premailer-cellpadding: 0;
      -premailer-cellspacing: 0;
    }
    /* Masthead ----------------------- */
    
    .email-masthead {
      padding: 25px 0;
      text-align: center;
    }
    
    .email-masthead_logo {
      width: 94px;
    }
    
    .email-masthead_name {
      font-size: 16px;
      font-weight: bold;
      color: #A8AAAF;
      text-decoration: none;
      text-shadow: 0 1px 0 white;
    }
    /* Body ------------------------------ */
    
    .email-body {
      width: 100%;
      margin: 0;
      padding: 0;
      -premailer-width: 100%;
      -premailer-cellpadding: 0;
      -premailer-cellspacing: 0;
    }
    
    .email-body_inner {
      width: 570px;
      margin: 0 auto;
      padding: 0;
      -premailer-width: 570px;
      -premailer-cellpadding: 0;
      -premailer-cellspacing: 0;
    }
    
    .email-footer {
      width: 570px;
      margin: 0 auto;
      padding: 0;
      -premailer-width: 570px;
      -premailer-cellpadding: 0;
      -premailer-cellspacing: 0;
      text-align: center;
    }
    
    .email-footer p {
      color: #A8AAAF;
    }
    
    .body-action {
      width: 100%;
      margin: 30px auto;
      padding: 0;
      -premailer-width: 100%;
      -premailer-cellpadding: 0;
      -premailer-cellspacing: 0;
      text-align: center;
    }
    
    .body-sub {
      margin-top: 25px;
      padding-top: 25px;
      border-top: 1px solid #EAEAEC;
    }
    
    .content-cell {
      padding: 35px;
    }
    /*Media Queries ------------------------------ */
    
    @media only screen and (max-width: 600px) {
      .email-body_inner,
      .email-footer {
        width: 100% !important;
      }
    }
    
    @media (prefers-color-scheme: dark) {
      body {
        background-color: #333333 !important;
        color: #FFF !important;
      }
      p,
      ul,
      ol,
      blockquote,
      h1,
      h2,
      h3,
      span,
      .purchase_item {
        color: #FFF !important;
      }
      .attributes_content,
      .discount {
        background-color: #222 !important;
      }
      .email-masthead_name {
        text-shadow: none !important;
      }
    }
    
    :root {
      color-scheme: light dark;
      supported-color-schemes: light dark;
    }
    </style>
    <!--[if mso]>
    <style type="text/css">
      .f-fallback  {
        font-family: Arial, sans-serif;
      }
    </style>
  <![endif]-->
    <style type="text/css" rel="stylesheet" media="all">
    body {
      width: 100% !important;
      height: 100%;
      margin: 0;
      -webkit-text-size-adjust: none;
    }
    
    body {
      font-family: "Nunito Sans", Helvetica, Arial, sans-serif;
    }
    
    body {
      background-color: #FFF;
      color: #333;
    }
    </style>
  </head>
  <body style="width: 100% !important; height: 100%; -webkit-text-size-adjust: none; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; background-color: #FFF; color: #333; margin: 0;" bgcolor="#FFF">
    <span class="preheader" style="display: none !important; visibility: hidden; mso-hide: all; font-size: 1px; line-height: 1px; max-height: 0; max-width: 0; opacity: 0; overflow: hidden;">This is a notification of your recent Goodrich order for {{.OrderDate}} {{.OrderTime}}.</span>
    <table class="email-wrapper" width="100%" cellpadding="0" cellspacing="0" role="presentation" style="width: 100%; -premailer-width: 100%; -premailer-cellpadding: 0; -premailer-cellspacing: 0; margin: 0; padding: 0;">
      <tr>
        <td align="center" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px;">
          <table class="email-content" width="100%" cellpadding="0" cellspacing="0" role="presentation" style="width: 100%; -premailer-width: 100%; -premailer-cellpadding: 0; -premailer-cellspacing: 0; margin: 0; padding: 0;">
            <tr>
              <td class="email-masthead" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px; text-align: center; padding: 25px 0;" align="center">
                <a href="https://wso.williams.edu/goodrich" class="f-fallback email-masthead_name" style="color: #A8AAAF; font-size: 16px; font-weight: bold; text-decoration: none; text-shadow: 0 1px 0 white;">
                WSO Goodrich Order
              </a>
              </td>
            </tr>
            <!-- Email Body -->
            <tr>
              <td class="email-body" width="570" cellpadding="0" cellspacing="0" style="word-break: break-word; margin: 0; padding: 0; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px; width: 100%; -premailer-width: 100%; -premailer-cellpadding: 0; -premailer-cellspacing: 0;">
                <table class="email-body_inner" align="center" width="570" cellpadding="0" cellspacing="0" role="presentation" style="width: 570px; -premailer-width: 570px; -premailer-cellpadding: 0; -premailer-cellspacing: 0; margin: 0 auto; padding: 0;">
                  <!-- Body content -->
                  <tr>
                    <td class="content-cell" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px; padding: 35px;">
                      <div class="f-fallback">
                        <h1 style="margin-top: 0; color: #333333; font-size: 22px; font-weight: bold; text-align: left;" align="left">Hi {{.User.Name}},</h1>
                        <p style="font-size: 16px; line-height: 1.625; color: #333; margin: .4em 0 1.1875em;">Thank you for ordering from Goodrich through the WSO service. This email is the receipt for your order.</p>
                        <p style="font-size: 16px; line-height: 1.625; color: #333; margin: .4em 0 1.1875em;">Your order will be ready at <b>{{.OrderTime}} on {{.OrderDate}}</b>.</p>
                        <p style="font-size: 16px; line-height: 1.625; color: #333; margin: .4em 0 1.1875em;">If you chose to pay with cash or credit card (or swipe + cash/card), you will need to pay up-front at Goodrich. Otherwise, if you chose to pay with only a meal swipe, your payment has already been processed. Your order can be picked up in the front area of Goodrich Hall.</p>
                        <table class="purchase" width="100%" cellpadding="0" cellspacing="0" role="presentation" style="width: 100%; -premailer-width: 100%; -premailer-cellpadding: 0; -premailer-cellspacing: 0; margin: 0; padding: 35px 0;">
                          <tr>
                            <td style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px;">
                              <h3 style="margin-top: 0; color: #333333; font-size: 14px; font-weight: bold; text-align: left;" align="left">Order #{{.Order.ID}}</h3></td>
                            <td style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px;">
                              <h3 class="align-right" style="margin-top: 0; color: #333333; font-size: 14px; font-weight: bold; text-align: right;" align="right">{{.TimeNow}}</h3></td>
                          </tr>
                          <tr>
                            <td style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px;">
                              <h3 style="margin-top: 0; color: #333333; font-size: 14px; font-weight: bold; text-align: left;" align="left">Payment:</h3></td>
                            <td style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px;">
                              <h3 class="align-right" style="margin-top: 0; color: #333333; font-size: 14px; font-weight: bold; text-align: right;" align="right">{{.PaymentString}}</h3></td>
                          </tr>
                          <tr>
                            <td colspan="2" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px;">
                              <table class="purchase_content" width="100%" cellpadding="0" cellspacing="0" style="width: 100%; -premailer-width: 100%; -premailer-cellpadding: 0; -premailer-cellspacing: 0; margin: 0; padding: 25px 0 0;">
                                <tr>
                                  <th class="purchase_heading" align="left" style="font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px; padding-bottom: 8px; border-bottom-width: 1px; border-bottom-color: #EAEAEC; border-bottom-style: solid;">
                                    <p class="f-fallback" style="font-size: 12px; line-height: 1.625; color: #85878E; margin: 0;">Item</p>
                                  </th>
                                  <th class="purchase_heading" align="right" style="font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px; padding-bottom: 8px; border-bottom-width: 1px; border-bottom-color: #EAEAEC; border-bottom-style: solid;">
                                    <p class="f-fallback" style="font-size: 12px; line-height: 1.625; color: #85878E; margin: 0;">Amount</p>
                                  </th>
                                </tr>
                                {{range .Order.Items}}
                                <tr>
                                  <td width="80%" class="purchase_item" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 15px; color: #51545E; line-height: 18px; padding: 10px 0;"><span class="f-fallback">{{.Item.Title}} {{if .Item.Type}}({{.Item.Type}}{{if .Note}}, {{.Note}}{{end}}){{end}}</span></td>
                                  <td class="align-right" width="20%" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px; text-align: right;" align="right"><span class="f-fallback">${{printf "%.2f" .Item.Price}}</span></td>
                                </tr>
                                {{end}}
								{{if .SwipeOrder}}
                                <tr>
                                  <td width="80%" class="purchase_footer" valign="middle" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px; padding-top: 15px; border-top-width: 1px; border-top-color: #EAEAEC; border-top-style: solid;">
                                    <p class="f-fallback purchase_total purchase_total--label" style="font-size: 16px; line-height: 1.625; text-align: right; font-weight: bold; color: #333333; margin: 0; padding: 0 15px 0 0;" align="right">Subtotal</p>
                                  </td>
                                  <td width="20%" class="purchase_footer" valign="middle" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px; padding-top: 15px; border-top-width: 1px; border-top-color: #EAEAEC; border-top-style: solid;">
                                    <p class="f-fallback purchase_total" style="font-size: 16px; line-height: 1.625; text-align: right; font-weight: bold; color: #333333; margin: 0;" align="right">${{printf "%.2f" .Order.TotalPrice}}</p>
                                  </td>
                                </tr>
                                <tr>
                                  <td width="80%" class="purchase_item" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 15px; color: #51545E; line-height: 18px; padding: 10px 0;"><span class="f-fallback">Meal Swipe</span></td>
                                  <td class="align-right" width="20%" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px; text-align: right;" align="right"><span class="f-fallback">($5.00)</span></td>
                                </tr>
								{{end}}
                                <tr>
                                  <td width="80%" class="purchase_footer" valign="middle" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px; padding-top: 15px; border-top-width: 1px; border-top-color: #EAEAEC; border-top-style: solid;">
                                    <p class="f-fallback purchase_total purchase_total--label" style="font-size: 16px; line-height: 1.625; text-align: right; font-weight: bold; color: #333333; margin: 0; padding: 0 15px 0 0;" align="right">Total Owed</p>
                                  </td>
                                  <td width="20%" class="purchase_footer" valign="middle" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px; padding-top: 15px; border-top-width: 1px; border-top-color: #EAEAEC; border-top-style: solid;">
                                    <p class="f-fallback purchase_total" style="font-size: 16px; line-height: 1.625; text-align: right; font-weight: bold; color: #333333; margin: 0;" align="right">${{printf "%.2f" .TotalPriceAdjusted}}</p>
                                  </td>
                                </tr>
                              </table>
                            </td>
                          </tr>
                        </table>
                        <p style="font-size: 16px; line-height: 1.625; color: #333; margin: .4em 0 1.1875em;">Please arrive only at your set time and leave once you have your order.</p>
                        <p style="font-size: 16px; line-height: 1.625; color: #333; margin: .4em 0 1.1875em;">Cheers,
                          <br />The WSO Goodrich Team</p>
                        <!-- Action -->
                        <table class="body-action" align="center" width="100%" cellpadding="0" cellspacing="0" role="presentation" style="width: 100%; -premailer-width: 100%; -premailer-cellpadding: 0; -premailer-cellspacing: 0; text-align: center; margin: 30px auto; padding: 0;">
                          <tr>
                            <td align="center" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px;">
                              <table width="100%" border="0" cellspacing="0" cellpadding="0" role="presentation">
                                <tr>
                                  <td align="center" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px;">
                                    <a href="https://wso.williams.edu/goodrich" class="f-fallback button button--blue" target="_blank" style="color: #FFF; border-color: #3869d4; border-style: solid; border-width: 10px 18px; background-color: #3869D4; display: inline-block; text-decoration: none; border-radius: 3px; box-shadow: 0 2px 3px rgba(0, 0, 0, 0.16); -webkit-text-size-adjust: none; box-sizing: border-box;">Check Your Order Online</a>
                                  </td>
                                </tr>
                              </table>
                            </td>
                          </tr>
                        </table>
                        <!-- Sub copy -->
                        <table class="body-sub" role="presentation" style="margin-top: 25px; padding-top: 25px; border-top-width: 1px; border-top-color: #EAEAEC; border-top-style: solid;">
                          <tr>
                            <td style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px;">
                              <p class="f-fallback sub" style="font-size: 13px; line-height: 1.625; color: #333; margin: .4em 0 1.1875em;">Interested in working for WSO? You can <a href="https://forms.gle/7EcorfSMSuLQw5XW8" style="color: #3869D4;">sign up here</a>.</p>
                            </td>
                          </tr>
                        </table>
                      </div>
                    </td>
                  </tr>
                </table>
              </td>
            </tr>
            <tr>
              <td style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px;">
                <table class="email-footer" align="center" width="570" cellpadding="0" cellspacing="0" role="presentation" style="width: 570px; -premailer-width: 570px; -premailer-cellpadding: 0; -premailer-cellspacing: 0; text-align: center; margin: 0 auto; padding: 0;">
                  <tr>
                    <td class="content-cell" align="center" style="word-break: break-word; font-family: &quot;Nunito Sans&quot;, Helvetica, Arial, sans-serif; font-size: 16px; padding: 35px;">
                      <p class="f-fallback sub align-center" style="font-size: 13px; line-height: 1.625; text-align: center; color: #A8AAAF; margin: .4em 0 1.1875em;" align="center">© 2021 Aidan Lloyd-Tucker &amp; WSO. All rights reserved.</p>
                    </td>
                  </tr>
                </table>
              </td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>`
