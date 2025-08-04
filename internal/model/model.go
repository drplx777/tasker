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
	DashboardName string     `json:"dashboardName,omitempty"`
	BlockedBy     []string   `db:"blockedBy" json:"blockedBy"`
}

type User struct {
	ID         int     `db:"id" json:"id"`
	Name       string  `db:"name" json:"name"`
	Surname    string  `db:"surname" json:"surname"`
	Middlename *string `db:"middlename" json:"middlename,omitempty"`
	Login      string  `db:"login" json:"login"`
	RoleID     int     `db:"roleID" json:"roleID"`
	Password   string  `db:"password" json:"-"`
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
