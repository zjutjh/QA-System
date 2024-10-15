package models

import "time"

type Survey struct {
	ID       int       `json:"id"`
	UserID   int       `json:"user_id"`  //用户id
	Title    string    `json:"title"`    //问卷标题
	Desc     string    `json:"desc"`     //问卷描述
	Img      string    `json:"img"`      //问卷图片
	Deadline time.Time `json:"deadline"` //截止时间
	Status   int       `json:"status"`   //问卷状态  1:未发布 2:已发布 3:已截止
	Num      int       `json:"num"`      //问卷填写数量
}
