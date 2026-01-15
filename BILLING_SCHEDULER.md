# Billing Scheduler

## Overview
Billing scheduler otomatis untuk membuat tagihan bulanan dan custom secara otomatis setiap tanggal 1 di bulan berjalan.

## Features

### 1. Automatic Monthly Billing Creation
- Berjalan otomatis setiap tanggal 1 jam 00:00:00
- Membuat billing bulanan (monthly) untuk semua user dengan role "penghuni"
- Menggunakan setting billing yang aktif (published)

### 2. Automatic Custom Billing Creation
- Berjalan bersamaan dengan monthly billing
- Membuat billing custom berdasarkan billing settings ID
- Default menggunakan billing_settings_id = 1 (dapat dikonfigurasi)

## Cron Schedule
```
"0 0 0 1 * *"
```
Format: `seconds minutes hours day-of-month month day-of-week`
- Berjalan pada detik ke-0, menit ke-0, jam ke-0 (midnight)
- Pada hari pertama setiap bulan
- Untuk semua bulan dan hari dalam seminggu

## Implementation

### File Structure
```
internal/
  scheduler/
    billing_scheduler.go    # Billing scheduler implementation
```

### Main Components

#### BillingScheduler struct
```go
type BillingScheduler struct {
    billingService service.BillingService
    logger         *logger.Logger
    cron           *cron.Cron
}
```

#### Key Methods
- `Start()` - Memulai scheduler dan mendaftarkan cron jobs
- `Stop()` - Menghentikan scheduler dengan graceful shutdown
- `RunNow()` - Menjalankan job secara manual (untuk testing)
- `createMonthlyBillings()` - Job utama yang dijalankan oleh scheduler

## Configuration

### Billing Settings ID
Untuk mengubah billing settings ID yang digunakan untuk custom billing, edit file `billing_scheduler.go`:

```go
// Di dalam method createMonthlyBillings()
customResponse, err := s.billingService.CreateBulkCustomBillingsForAllUsers(month, 1, year)
//                                                                                  ^
//                                                                  Ubah ID di sini
```

### Multiple Custom Billing Settings
Jika Anda memiliki multiple custom billing settings yang perlu dijalankan, Anda dapat:

1. Tambahkan method di repository untuk mengambil semua active custom billing settings
2. Loop melalui semua settings dan panggil `CreateBulkCustomBillingsForAllUsers` untuk masing-masing

Contoh:
```go
// Get all active custom billing settings
settings, err := s.billingRepo.GetActiveCustomSettingBillings()
if err != nil {
    s.logger.WithField("error", err).Error("Failed to get custom billing settings")
    return
}

// Create custom billings for each setting
for _, setting := range settings {
    customResponse, err := s.billingService.CreateBulkCustomBillingsForAllUsers(month, int(setting.ID), year)
    if err != nil {
        s.logger.WithField("error", err).WithField("setting_id", setting.ID).Error("Failed to create custom billings")
    } else {
        s.logger.WithField("response", customResponse).WithField("setting_id", setting.ID).Info("Custom billings created successfully")
    }
}
```

## Logging

Scheduler secara otomatis mencatat semua aktivitas ke tabel `log_schedullers` dengan status berikut:

### Status Codes:
- **START**: Ketika scheduler mulai dijalankan
- **RUNNING**: Ketika proses billing sedang berjalan
- **FAILED**: Jika terjadi error selama proses
- **SUCCESS**: Jika billing berhasil dibuat

### Log Fields:
- `scheduller_code`: "MONTHLY_BILLING_CREATION"
- `document_id`: UUID unik untuk setiap eksekusi
- `message`: Detail pesan dan response dari operasi
- `status_scheduller`: Status eksekusi (START/RUNNING/FAILED/SUCCESS)
- `created_at`, `updated_at`, `published_at`: Timestamp
- `created_by_id`, `updated_by_id`: Admin ID (1)
- `locale`: "en"

### Contoh Log Entries:
```sql
-- START log
INSERT INTO log_schedullers (document_id, scheduller_code, message, status_scheduller, created_at, created_by_id)
VALUES ('uuid-123', 'MONTHLY_BILLING_CREATION', 'Starting scheduled monthly billing creation', 'START', NOW(), 1);

-- SUCCESS log with response
INSERT INTO log_schedullers (document_id, scheduller_code, message, status_scheduller, created_at, created_by_id)
VALUES ('uuid-123', 'MONTHLY_BILLING_CREATION', '{"total_users":50,"total_billings":150,"success_count":150,"failed_count":0}', 'SUCCESS', NOW(), 1);
```

## Testing

### Manual Testing
Untuk testing tanpa menunggu tanggal 1, Anda dapat:

1. **Menjalankan job secara manual** (tambahkan endpoint):
```go
// Di handler
func (h *BillingHandler) RunSchedulerNow(c *gin.Context) {
    // Call scheduler RunNow method
    // Implementasi tergantung cara Anda menyimpan scheduler instance
}
```

2. **Mengubah cron schedule untuk testing**:
```go
// Ubah dari "0 0 0 1 * *" menjadi (contoh: setiap menit)
_, err := s.cron.AddFunc("0 * * * * *", s.createMonthlyBillings)
```

3. **Test secara langsung** dengan memanggil service method:
```go
month := int(time.Now().Month())
year := time.Now().Year()
response, err := billingService.CreateBulkMonthlyBillingsForAllUsers(month, year)
```

## Error Handling

Scheduler dirancang untuk:
- ✅ Mencatat error tetapi tetap berjalan
- ✅ Tidak crash aplikasi jika billing creation gagal
- ✅ Memberikan informasi detail di log untuk debugging
- ✅ Melanjutkan ke custom billing meskipun monthly billing gagal

## Integration with Main Application

Scheduler diinisialisasi di `cmd/server/main.go`:

```go
// Initialize and start billing scheduler
billingScheduler := scheduler.NewBillingScheduler(billingService, appLogger)
if err := billingScheduler.Start(); err != nil {
    appLogger.WithField("error", err).Fatal("Failed to start billing scheduler")
}

// Stop scheduler saat aplikasi shutdown
defer billingScheduler.Stop()
```

## Dependencies

- `github.com/robfig/cron/v3` - Library cron scheduler
- BillingService - Service layer untuk billing operations
- Logger - Aplikasi logger untuk logging

## Maintenance

### Monitoring
Pastikan untuk memonitor:
- Log aplikasi pada tanggal 1 setiap bulan
- Success rate dari billing creation
- Error patterns jika ada

### Common Issues

1. **Billing tidak terbuat**
   - Cek apakah ada setting billing yang active
   - Cek apakah ada user dengan role "penghuni"
   - Lihat log untuk error details

2. **Duplicate billing**
   - Pastikan tidak ada multiple instance aplikasi yang berjalan
   - Cek database untuk duplicate entries

3. **Performance issues**
   - Jika user sangat banyak, pertimbangkan untuk menjalankan dalam batch
   - Monitor memory usage saat job berjalan

## Future Improvements

Potential enhancements:
- [ ] Konfigurasi schedule via database atau config file
- [ ] Web UI untuk monitoring scheduler status
- [ ] Notification (email/slack) setelah job selesai
- [ ] Retry mechanism untuk failed billing creation
- [ ] Dashboard untuk melihat scheduler history
- [ ] Support untuk multiple schedules (misal: reminder sebelum jatuh tempo)
