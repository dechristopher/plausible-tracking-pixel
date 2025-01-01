package internal

import (
	// stdlib
	"bytes"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	// third-party
	"github.com/gofiber/fiber/v2"
)

type PlausibleEvent string

const PageViewType PlausibleEvent = "pageview"

type PixelController struct {
	EventEndpoint   string
	VisitorIPHeader string
	FS              embed.FS

	pixelCache []byte
}

func (pc *PixelController) Handler(ctx *fiber.Ctx) error {
	page := ctx.Query("p")

	if page == "" {
		log.Printf("dchr/ptp: no page provided in query string: %s", pc.EventEndpoint)
		return fiber.ErrBadRequest
	}

	userAgent := ctx.Get("User-Agent", "unknown")
	referer := ctx.Get("Referer")

	parsedUrl, errParseUrl := url.Parse(referer)
	if errParseUrl != nil {
		log.Printf("dchr/ptp: error parsing referer: %s", errParseUrl)
		return fiber.ErrBadRequest
	}

	builtUrl, errAddPage := parsedUrl.Parse(page)
	if errAddPage != nil {
		log.Printf("dchr/ptp: error appending page: %s", errAddPage)
		return fiber.ErrBadRequest
	}

	// store pixel in-memory
	if pc.pixelCache == nil {
		pixel, errRead := pc.FS.ReadFile("static/pixel.png")
		if errRead != nil {
			log.Printf("dchr/ptp: error reading pixel.png: %s", errRead)
			return fiber.ErrInternalServerError
		}

		pc.pixelCache = pixel
	}

	// immediately dispatch event to Plausible endpoint
	go pc.sendPlausibleEvent(plausibleEvent{
		UserAgent:     userAgent,
		XForwardedFor: ctx.Get(pc.VisitorIPHeader),
		EventName:     string(PageViewType),
		URL:           builtUrl.String(),
		Domain:        builtUrl.Host,
	})

	ctx.Set("Content-Type", "image/png")
	_, errWrite := ctx.Write(pc.pixelCache)
	if errWrite != nil {
		log.Printf("dchr/ptp: failed to write pixel: %s", errWrite)
		return fiber.ErrInternalServerError
	}

	return nil
}

func (pc *PixelController) sendPlausibleEvent(e plausibleEvent) {
	log.Printf("dchr/ptp: sending '%s' event: %+v", e.EventName, e)

	body, errMarshal := json.Marshal(e)
	if errMarshal != nil {
		log.Printf("dchr/ptp: failed to marshal event: %s, %+v", errMarshal, e)
		return
	}

	req, errReq := http.NewRequest(http.MethodPost, pc.EventEndpoint, bytes.NewBuffer(body))
	if errReq != nil {
		panic(errReq)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", e.UserAgent)
	req.Header.Set("X-Forwarded-For", e.XForwardedFor)

	resp, errDo := http.DefaultClient.Do(req)
	if errDo != nil {
		log.Printf("dchr/ptp: failed to send event: %s, %+v", errDo, e)
		return
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		log.Printf("dchr/ptp: non-20X received when sending event: %d %+v", resp.StatusCode, e)
		return
	}

	log.Printf("dchr/ptp: '%s' event sent: %s", e.EventName, e.URL)
}
