package model

import "time"

type Task struct {
	ID            string     `db:"id" json:"id"`
	Title         string     `db:"title" json:"title"`
	Description   string     `db:"description" json:"description"`
	Status        string     `db:"status" json:"status"`
	ReporterID    string     `db:"reporterD" json:"reporterId"`
	AssignerID    *string    `db:"assignerID" json:"assignerId,omitempty"`
	ReviewerID    *string    `db:"reviewerID" json:"reviewerId,omitempty"`
	ApproverID    string     `db:"approverID" json:"approverId"`
	ApproveStatus string     `db:"approveStatus" json:"approveStatus"`
	CreatedAt     time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updatedAt"`
	StartedAt     *time.Time `db:"started_At" json:"startedAt,omitempty"`
	CompletedAt   *time.Time `db:"done_at" json:"completedAt,omitempty"`
	DeadLine      string     `db:"deadline" json:"deadline"`
	DashboardID   string     `db:"dashboardID" json:"dashboardId"`
	BlockedBy     []string   `db:"blockedBy" json:"blockedBy"`
	Space         *string    `db:"space" json:"space,omitempty"`
	AssignerName  *string    `json:"assignerName,omitempty"`
	ApproverName  *string    `json:"approverName,omitempty"`
	ReporterName  *string    `json:"reporterName,omitempty"`
	DashboardName *string    `json:"dashboardName,omitempty"`
}

type TaskPatch struct {
	Title         *string    `json:"title,omitempty"`
	Description   *string    `json:"description,omitempty"`
	Status        *string    `json:"status,omitempty"`
	ReporterID    *string    `json:"reporterId,omitempty"`
	AssignerID    *string    `json:"assignerId,omitempty"`
	ReviewerID    *string    `json:"reviewerId,omitempty"`
	ApproverID    *string    `json:"approverId,omitempty"`
	ApproveStatus *string    `json:"approveStatus,omitempty"`
	StartedAt     *time.Time `json:"startedAt,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	DeadLine      *time.Time `json:"deadline,omitempty"`
	DashboardID   *string    `json:"dashboardId,omitempty"`
	BlockedBy     *[]string  `json:"blockedBy,omitempty"`
}

type User struct {
	ID         int       `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	Surname    string    `db:"surname" json:"surname"`
	Middlename *string   `db:"middlename" json:"middlename,omitempty"`
	Login      string    `db:"login" json:"login"`
	RoleID     int       `db:"roleID" json:"roleID"`
	Password   string    `db:"password" json:"-"`
	Spaces     *[]string `db:"spaces" json:"spaces,omitempty"`
}

type RegisterRequest struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Middlename string `json:"middlename"`
	Login      string `json:"login"`
	RoleID     int    `json:"roleID"`
	Password   string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
type DashBoards struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
