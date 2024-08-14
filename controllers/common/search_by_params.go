package common

import (
	"github.com/goravel/framework/contracts/database/orm"
)

type SearchByParamsService struct {
}

func (s *SearchByParamsService) SearchByParams(params map[string]string,excepts ...string) func(methods orm.Query) orm.Query {
	for _, except := range excepts {
		delete(params, except)
	}

	return func(query orm.Query) orm.Query {
		//从params的key中将excepts的key去除
		for key, value := range params {
			if key == "pappt_name" && value != "" {
				continue
			}

			if value == "" || key == "pageSize" || key == "total" || key == "currentPage" || key == "sort" || key == "order" {
				continue
			} else {
				query = query.Where(key+" like ?", "%"+value+"%")
			}
		}
		return query
	}
}
