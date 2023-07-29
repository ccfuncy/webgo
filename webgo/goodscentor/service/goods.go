package service

import "goodscentor/model"

type GoodsService struct {
}

func (s *GoodsService) Find(id int64) *model.Result {
	goods := model.Goods{
		Id:   9002,
		Name: "商品9002",
	}
	return &model.Result{
		Code: 200,
		Msg:  "success",
		Data: goods,
	}
}
