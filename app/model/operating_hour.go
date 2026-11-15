package model

// OperatingHour adalah jam buka perpustakaan untuk satu hari.
// DayOfWeek: 0 = Minggu, 1 = Senin, ..., 6 = Sabtu.
// OpenTime dan CloseTime berformat "HH:MM", contoh "08:00".
type OperatingHour struct {
	ID        int    `json:"id"`
	DayOfWeek int    `json:"day_of_week"`
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
	IsOpen    bool   `json:"is_open"`
}

// UpdateOperatingHourRequest adalah body untuk PUT /operating-hours/:day (admin).
type UpdateOperatingHourRequest struct {
	OpenTime  string `json:"open_time" validate:"required,datetime=15:04"`
	CloseTime string `json:"close_time" validate:"required,datetime=15:04"`
	IsOpen    bool   `json:"is_open"`
}
