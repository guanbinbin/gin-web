package cache_service

import (
	"gin-web/models"
	"gin-web/pkg/global"
	"gin-web/pkg/service"
	"gin-web/pkg/utils"
)

// 获取权限菜单树
func (s *RedisService) GetMenuTree(roleId uint) ([]models.SysMenu, error) {
	if !global.Conf.System.UseRedis {
		// 不使用redis
		return s.mysql.GetMenuTree(roleId)
	}
	tree := make([]models.SysMenu, 0)
	var role models.SysRole
	err := s.GetItemByIdFromCache(roleId, &role, role.TableName())
	if err != nil {
		return tree, err
	}
	// 当前角色拥有的全部菜单, 生成菜单树
	tree = service.GenMenuTree(nil, s.getMenusByRoleId(roleId))
	return tree, nil
}

// 获取所有菜单
func (s *RedisService) GetMenus() ([]models.SysMenu, error) {
	if !global.Conf.System.UseRedis {
		// 不使用redis
		return s.mysql.GetMenus()
	}
	tree := make([]models.SysMenu, 0)
	menus := make([]models.SysMenu, 0)
	jsonMenus := s.GetListFromCache(nil, new(models.SysMenu).TableName())
	// 查询所有菜单, 根据sort字段排序
	res := s.JsonQuery().FromString(jsonMenus).SortBy("sort").Get()
	// 转换为结构体
	utils.Struct2StructByJson(res, &menus)

	// 生成菜单树
	tree = service.GenMenuTree(nil, menus)
	return tree, nil
}

// 根据权限编号获取全部菜单
func (s *RedisService) GetAllMenuByRoleId(roleId uint) ([]models.SysMenu, []uint, error) {
	if !global.Conf.System.UseRedis {
		// 不使用redis
		return s.mysql.GetAllMenuByRoleId(roleId)
	}
	// 菜单树
	tree := make([]models.SysMenu, 0)
	// 有权限访问的id列表
	accessIds := make([]uint, 0)
	allMenu, _ := s.GetMenus()
	// 查询角色拥有的全部菜单
	roleMenus := s.getMenusByRoleId(roleId)
	// 生成菜单树
	tree = service.GenMenuTree(nil, allMenu)
	// 获取id列表
	for _, menu := range roleMenus {
		accessIds = append(accessIds, menu.Id)
	}
	return tree, accessIds, nil
}

// 获取当前角色拥有的全部菜单
func (s *RedisService) getMenusByRoleId(roleId uint) []models.SysMenu {
	menus := make([]models.SysMenu, 0)
	relations := make([]models.RelationRoleMenu, 0)
	_ = s.GetListFromCache(&relations, new(models.RelationRoleMenu).TableName())
	jsonMenus := s.GetListFromCache(nil, new(models.SysMenu).TableName())
	// JsonQuery只支持int数组, 不支持uint
	menuIds := make([]int, 0)
	for _, relation := range relations {
		if relation.SysRoleId == roleId {
			menuIds = append(menuIds, int(relation.SysMenuId))
		}
	}
	res := s.JsonQuery().FromString(jsonMenus).WhereIn("id", menuIds).Get()

	// 转换为结构体
	utils.Struct2StructByJson(res, &menus)
	return menus
}
