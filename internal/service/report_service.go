package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nicolaananda/catatuang/internal/domain"
	"github.com/nicolaananda/catatuang/internal/repository"
)

type ReportService struct {
	txRepo *repository.TransactionRepository
}

func NewReportService(txRepo *repository.TransactionRepository) *ReportService {
	return &ReportService{txRepo: txRepo}
}

type ReportSummary struct {
	TotalIncome   float64
	TotalExpense  float64
	NetBalance    float64
	TopCategories map[string]float64
	Transactions  []*domain.Transaction
}

func (s *ReportService) GenerateReport(ctx context.Context, userID int64, start, end time.Time) (*ReportSummary, error) {
	transactions, err := s.txRepo.GetByUserAndDateRange(ctx, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	summary := &ReportSummary{
		TopCategories: make(map[string]float64),
		Transactions:  transactions,
	}

	for _, tx := range transactions {
		if tx.Type == domain.TypeIncome {
			summary.TotalIncome += tx.Amount
		} else {
			summary.TotalExpense += tx.Amount
		}

		// Aggregate by category
		if tx.Category != "" {
			summary.TopCategories[tx.Category] += tx.Amount
		}
	}

	summary.NetBalance = summary.TotalIncome - summary.TotalExpense

	return summary, nil
}

func (s *ReportService) FormatReport(summary *ReportSummary, period string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📊 *Rekap %s*\n\n", period))
	sb.WriteString(fmt.Sprintf("💰 Pemasukan: Rp %.0f\n", summary.TotalIncome))
	sb.WriteString(fmt.Sprintf("💸 Pengeluaran: Rp %.0f\n", summary.TotalExpense))
	sb.WriteString(fmt.Sprintf("📈 Saldo Bersih: Rp %.0f\n\n", summary.NetBalance))

	if len(summary.TopCategories) > 0 {
		sb.WriteString("🏷️ *Top Kategori:*\n")
		for cat, amount := range summary.TopCategories {
			sb.WriteString(fmt.Sprintf("  • %s: Rp %.0f\n", cat, amount))
		}
	}

	if len(summary.Transactions) == 0 {
		sb.WriteString("\nBelum ada transaksi di periode ini.")
	}

	return sb.String()
}

func (s *ReportService) GetDailyReport(ctx context.Context, userID int64, loc *time.Location) (string, error) {
	now := time.Now().In(loc)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	end := start.Add(24 * time.Hour)

	summary, err := s.GenerateReport(ctx, userID, start, end)
	if err != nil {
		return "", err
	}

	return s.FormatReport(summary, "Hari Ini"), nil
}

func (s *ReportService) GetWeeklyReport(ctx context.Context, userID int64, loc *time.Location) (string, error) {
	now := time.Now().In(loc)
	// Start of week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := now.AddDate(0, 0, -(weekday - 1))
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	end := start.AddDate(0, 0, 7)

	summary, err := s.GenerateReport(ctx, userID, start, end)
	if err != nil {
		return "", err
	}

	return s.FormatReport(summary, "Minggu Ini"), nil
}

func (s *ReportService) GetMonthlyReport(ctx context.Context, userID int64, loc *time.Location) (string, error) {
	now := time.Now().In(loc)
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, 0)

	summary, err := s.GenerateReport(ctx, userID, start, end)
	if err != nil {
		return "", err
	}

	return s.FormatReport(summary, "Bulan Ini"), nil
}
