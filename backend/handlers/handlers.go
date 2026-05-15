package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fleetify/models"
	"fleetify/repository"
	"fleetify/service"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	userRepo    *repository.UserRepository
	vehicleRepo *repository.VehicleRepository
	itemRepo    *repository.MasterItemRepository
	reportRepo  *repository.ReportRepository
	webhook     *service.WebhookService
	uploadDir   string
}

func NewHandler(
	userRepo *repository.UserRepository,
	vehicleRepo *repository.VehicleRepository,
	itemRepo *repository.MasterItemRepository,
	reportRepo *repository.ReportRepository,
	webhook *service.WebhookService,
	uploadDir string,
) *Handler {
	return &Handler{
		userRepo:    userRepo,
		vehicleRepo: vehicleRepo,
		itemRepo:    itemRepo,
		reportRepo:  reportRepo,
		webhook:     webhook,
		uploadDir:   uploadDir,
	}
}

func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	users, err := h.userRepo.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(users)
}

func (h *Handler) ListVehicles(c *fiber.Ctx) error {
	vehicles, err := h.vehicleRepo.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(vehicles)
}

func (h *Handler) ListMasterItems(c *fiber.Ctx) error {
	items, err := h.itemRepo.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

type createReportItemReq struct {
	ItemID   uint64  `json:"item_id"`
	ItemName string  `json:"item_name"`
	Type     string  `json:"type"`
	Price    float64 `json:"price"`
	Quantity uint    `json:"quantity"`
}

type createReportReq struct {
	VehicleID    uint64                `json:"vehicle_id"`
	LicensePlate string                `json:"license_plate"`
	VehicleModel string                `json:"vehicle_model"`
	Odometer     uint                  `json:"odometer"`
	Complaint    string                `json:"complaint"`
	InitialPhoto string                `json:"initial_photo"`
	Items        []createReportItemReq `json:"items"`
}

func (h *Handler) CreateReport(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	req, err := h.parseCreateReport(c)
	if err != nil {
		log.Printf("[ERROR] parseCreateReport: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if strings.TrimSpace(req.LicensePlate) == "" && req.VehicleID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "license_plate is required"})
	}
	if req.Odometer == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "odometer is required"})
	}
	if strings.TrimSpace(req.Complaint) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "complaint is required"})
	}
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "at least one item is required"})
	}

	vehicleID, err := h.resolveVehicleID(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	reportItems, priceMap, err := h.resolveReportItems(req.Items)
	if err != nil {
		log.Printf("[ERROR] resolveReportItems: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var initialPhoto *string
	if file, err := c.FormFile("initial_photo"); err == nil {
		path, saveErr := h.saveUpload(file, "initial")
		if saveErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": saveErr.Error()})
		}
		initialPhoto = &path
	} else if req.InitialPhoto != "" {
		initialPhoto = &req.InitialPhoto
	} else {
		sim := fmt.Sprintf("/uploads/simulated_initial_%d.jpg", time.Now().UnixNano())
		initialPhoto = &sim
	}

	input := repository.CreateReportInput{
		VehicleID:    vehicleID,
		CreatedBy:    user.ID,
		Odometer:     req.Odometer,
		Complaint:    req.Complaint,
		InitialPhoto: initialPhoto,
		Items:        reportItems,
	}

	report, err := h.reportRepo.Create(input, priceMap)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(report)
}

func (h *Handler) parseCreateReport(c *fiber.Ctx) (createReportReq, error) {
	var req createReportReq
	contentType := string(c.Request().Header.ContentType())

	if strings.Contains(contentType, "multipart/form-data") {
		req.VehicleID = parseUint(formValue(c, "vehicle_id"))
		req.LicensePlate = formValue(c, "license_plate")
		req.VehicleModel = formValue(c, "vehicle_model")
		req.Odometer = uint(parseUint(formValue(c, "odometer")))
		req.Complaint = formValue(c, "complaint")
		req.InitialPhoto = formValue(c, "initial_photo")
		itemsRaw := formValue(c, "items")
		if itemsRaw == "" {
			return req, fmt.Errorf("items field is required (empty or missing)")
		}
		if err := json.Unmarshal([]byte(itemsRaw), &req.Items); err != nil {
			return req, fmt.Errorf("invalid items JSON: %v (got: %s)", err, itemsRaw)
		}
		return req, nil
	}

	if err := c.BodyParser(&req); err != nil {
		return req, fmt.Errorf("invalid request body: %v", err)
	}
	return req, nil
}

func (h *Handler) ListReports(c *fiber.Ctx) error {
	reports, err := h.reportRepo.ListAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(reports)
}

func (h *Handler) GetReport(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	report, err := h.reportRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrReportNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "report not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(report)
}

func (h *Handler) ApproveReport(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	report, err := h.reportRepo.Approve(id)
	if err != nil {
		if errors.Is(err, repository.ErrReportNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "report not found"})
		}
		if errors.Is(err, repository.ErrInvalidStatusChange) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "report must be PENDING_APPROVAL to approve"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	plate := ""
	if report.Vehicle.LicensePlate != "" {
		plate = report.Vehicle.LicensePlate
	}
	h.webhook.NotifyStatusChange(report.ID, report.Status, plate)
	return c.JSON(report)
}

func (h *Handler) CompleteReport(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	file, err := c.FormFile("proof_photo")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "proof_photo is required"})
	}
	proofPath, err := h.saveUpload(file, "proof")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	report, err := h.reportRepo.Complete(id, proofPath)
	if err != nil {
		if errors.Is(err, repository.ErrReportNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "report not found"})
		}
		if errors.Is(err, repository.ErrInvalidStatusChange) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "report must be APPROVED to complete"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	plate := ""
	if report.Vehicle.LicensePlate != "" {
		plate = report.Vehicle.LicensePlate
	}
	h.webhook.NotifyStatusChange(report.ID, report.Status, plate)
	return c.JSON(report)
}

func (h *Handler) saveUpload(file *multipart.FileHeader, prefix string) (string, error) {
	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		return "", err
	}
	ext := filepath.Ext(file.Filename)
	name := fmt.Sprintf("%s_%d%s", prefix, time.Now().UnixNano(), ext)
	dest := filepath.Join(h.uploadDir, name)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	out, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}

func (h *Handler) resolveReportItems(items []createReportItemReq) ([]repository.ReportItemInput, map[uint64]float64, error) {
	reportItems := make([]repository.ReportItemInput, 0, len(items))
	priceMap := make(map[uint64]float64)

	for idx, it := range items {
		if it.Quantity == 0 {
			return nil, nil, fmt.Errorf("item[%d]: quantity must be > 0", idx)
		}

		if strings.TrimSpace(it.ItemName) != "" {
			itemType := models.ItemType(strings.ToUpper(strings.TrimSpace(it.Type)))
			if itemType != models.ItemPart && itemType != models.ItemService {
				return nil, nil, fmt.Errorf("item[%d]: invalid type '%s'", idx, it.Type)
			}
			if it.Price <= 0 {
				return nil, nil, fmt.Errorf("item[%d]: price must be > 0", idx)
			}
			master, err := h.itemRepo.FindOrCreate(it.ItemName, itemType, it.Price)
			if err != nil {
				return nil, nil, fmt.Errorf("item[%d] FindOrCreate error: %v", idx, err)
			}
			snapshot := it.Price
			if snapshot <= 0 {
				snapshot = master.Price
			}
			reportItems = append(reportItems, repository.ReportItemInput{
				ItemID:   master.ID,
				Quantity: it.Quantity,
			})
			priceMap[master.ID] = snapshot
			continue
		}

		if it.ItemID == 0 {
			return nil, nil, fmt.Errorf("item[%d]: must have item_id or item_name", idx)
		}

		masterItems, err := h.itemRepo.FindByIDs([]uint64{it.ItemID})
		if err != nil {
			return nil, nil, fmt.Errorf("item[%d] FindByIDs error: %v", idx, err)
		}
		if len(masterItems) == 0 {
			return nil, nil, fmt.Errorf("item[%d]: master item with id=%d not found", idx, it.ItemID)
		}
		reportItems = append(reportItems, repository.ReportItemInput{
			ItemID:   it.ItemID,
			Quantity: it.Quantity,
		})
		priceMap[it.ItemID] = masterItems[0].Price
	}

	return reportItems, priceMap, nil
}

func (h *Handler) resolveVehicleID(req createReportReq) (uint64, error) {
	plate := strings.TrimSpace(req.LicensePlate)
	if plate != "" {
		vehicle, err := h.vehicleRepo.FindOrCreate(plate, req.VehicleModel)
		if err != nil {
			return 0, err
		}
		return vehicle.ID, nil
	}
	if req.VehicleID == 0 {
		return 0, fmt.Errorf("license_plate or vehicle_id is required")
	}
	if _, err := h.vehicleRepo.FindByID(req.VehicleID); err != nil {
		return 0, fmt.Errorf("vehicle not found")
	}
	return req.VehicleID, nil
}

func formValue(c *fiber.Ctx, key string) string {
	if v := c.FormValue(key); v != "" {
		return v
	}
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return ""
	}
	vals, ok := form.Value[key]
	if !ok || len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func parseIDParam(c *fiber.Ctx) (uint64, error) {
	return strconv.ParseUint(c.Params("id"), 10, 64)
}

func parseUint(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}
