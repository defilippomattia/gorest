package companies

type Company struct {
	ID          int    `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	YearFounded int    `json:"year_founded" validate:"required"`
}

type CompanyRequest struct {
	Name        string `json:"name" validate:"required"`
	YearFounded int    `json:"year_founded" validate:"required"`
}
