package dbs

import (
	"sync"

	"github.com/limes-cloud/kratosx"
	"gorm.io/gorm"

	"github.com/limes-cloud/manager/internal/domain/entity"
	"github.com/limes-cloud/manager/internal/types"
)

type User struct {
}

var (
	userIns  *User
	userOnce sync.Once
)

func NewUser() *User {
	userOnce.Do(func() {
		userIns = &User{}
	})
	return userIns
}

// GetUserByPhone 获取指定数据
func (infra *User) GetUserByPhone(ctx kratosx.Context, phone string) (*entity.User, error) {
	var user entity.User
	db := ctx.DB().Preload("Roles").Preload("Jobs").Preload("Departments")
	return &user, db.Where("phone = ?", phone).First(&user).Error
}

// GetUserByEmail 获取指定数据
func (infra *User) GetUserByEmail(ctx kratosx.Context, email string) (*entity.User, error) {
	var user entity.User
	db := ctx.DB().Preload("Roles").Preload("Jobs").Preload("Departments")
	return &user, db.Where("email = ?", email).First(&user).Error
}

// GetUser 获取指定的数据
func (infra *User) GetUser(ctx kratosx.Context, id uint32) (*entity.User, error) {
	var user entity.User
	db := ctx.DB().Preload("Roles").Preload("Jobs").Preload("Departments")
	return &user, db.First(&user, id).Error
}

// GetBaseUser 获取指定的数据
func (infra *User) GetBaseUser(ctx kratosx.Context, id uint32) (*entity.User, error) {
	var user entity.User
	return &user, ctx.DB().First(&user, id).Error
}

// ListUser 获取列表 fixed code
func (infra *User) ListUser(ctx kratosx.Context, req *types.ListUserRequest) ([]*entity.User, uint32, error) {
	var (
		total int64
		fs    = []string{"*"}
		list  []*entity.User
	)

	db := ctx.DB().Model(entity.User{}).Select(fs).
		Preload("Roles").
		Preload("Departments").
		Preload("Jobs")
	if req.ID != nil {
		db = db.Where("id = ?", *req.ID)
	}

	if req.Name != nil {
		db = db.Where("name LIKE ?", *req.Name+"%")
	}
	if req.Phone != nil {
		db = db.Where("phone = ?", *req.Phone)
	}
	if req.Email != nil {
		db = db.Where("email = ?", *req.Email)
	}
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}
	if len(req.LoggedAts) == 2 {
		db = db.Where("logged_at BETWEEN ? AND ?", req.LoggedAts[0], req.LoggedAts[1])
	}
	if len(req.CreatedAts) == 2 {
		db = db.Where("created_at BETWEEN ? AND ?", req.CreatedAts[0], req.CreatedAts[1])
	}
	if req.DepartmentIds != nil {
		db = db.Where("id in (?)",
			ctx.DB().Select("user_id").
				Model(entity.UserDepartment{}).
				Where("department_id in ?", req.DepartmentIds),
		)
	}
	// if req.RoleIds != nil {
	//	db = db.Where("role_id = ?", *req.RoleIds)
	// }

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return list, uint32(total), db.Find(&list).Error
}

// CreateUser 创建数据
func (infra *User) CreateUser(ctx kratosx.Context, user *entity.User) (uint32, error) {
	return user.Id, ctx.DB().Create(user).Error
}

// UpdateUser 更新数据
func (infra *User) UpdateUser(ctx kratosx.Context, user *entity.User) error {
	return ctx.DB().Transaction(func(tx *gorm.DB) error {
		if len(user.UserDepartments) != 0 {
			if err := tx.Where("user_id=?", user.Id).Delete(entity.UserDepartment{}).Error; err != nil {
				return err
			}
		}
		if len(user.UserRoles) != 0 {
			if err := tx.Where("user_id=?", user.Id).Delete(entity.UserRole{}).Error; err != nil {
				return err
			}
		}
		if len(user.UserJobs) != 0 {
			if err := tx.Where("user_id=?", user.Id).Delete(entity.UserJob{}).Error; err != nil {
				return err
			}
		}
		return tx.Updates(user).Error
	})
}

// ForceUpdateUser 强制更新数据
func (infra *User) ForceUpdateUser(ctx kratosx.Context, user *entity.User) error {
	return ctx.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id=?", user.Id).Delete(entity.UserDepartment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id=?", user.Id).Delete(entity.UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id=?", user.Id).Delete(entity.UserJob{}).Error; err != nil {
			return err
		}
		return tx.Updates(user).Error
	})
}

// UpdateUserStatus 更新数据状态
func (infra *User) UpdateUserStatus(ctx kratosx.Context, id uint32, status bool) error {
	return ctx.DB().Model(entity.User{}).Where("id=?", id).Update("status", status).Error
}

// DeleteUser 删除数据
func (infra *User) DeleteUser(ctx kratosx.Context, id uint32) error {
	return ctx.DB().Where("id = ?", id).Delete(&entity.User{}).Error
}

func (infra *User) GetUserRoleIds(ctx kratosx.Context, id uint32) ([]uint32, error) {
	var ids []uint32
	return ids, ctx.DB().Model(entity.UserRole{}).Where("user_id=?", id).Scan(&ids).Error
}

// ListLoginLog 获取列表 fixed code
func (infra *User) ListLoginLog(ctx kratosx.Context, req *types.ListLoginLogRequest) ([]*entity.LoginLog, uint32, error) {
	var (
		total int64
		fs    = []string{"*"}
		list  []*entity.LoginLog
	)

	db := ctx.DB().Model(entity.LoginLog{}).Select(fs)

	if req.Username != nil {
		db = db.Where("username = ?", *req.Username)
	}
	if len(req.CreatedAts) == 2 {
		db = db.Where("created_at BETWEEN ? AND ?", req.CreatedAts[0], req.CreatedAts[1])
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	db = db.Order("id desc")
	db = db.Offset(int((req.Page - 1) * req.PageSize)).Limit(int(req.PageSize))
	return list, uint32(total), db.Find(&list).Error
}

// CreateLoginLog 创建数据
func (infra *User) CreateLoginLog(ctx kratosx.Context, loginLog *entity.LoginLog) (uint32, error) {
	return loginLog.Id, ctx.DB().Create(loginLog).Error
}
