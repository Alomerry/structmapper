package copier

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestCopyModelToModel(t *testing.T) {
	type OriginLocation struct {
		City      string
		Latitude  float64
		Longitude float64
	}
	type originModel struct {
		Name      string
		Age       int
		Money     float64
		BirthDay  time.Time
		IsDeleted bool
		Location  OriginLocation
		Detail    *OriginLocation
	}

	var target originModel
	origin := originModel{
		Name:      "MockModel",
		Age:       21,
		Money:     7.7,
		BirthDay:  time.Now(),
		IsDeleted: false,
		Location: OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
		Detail: &OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
	}

	assert.Nil(t, Instance(nil).From(origin).CopyTo(&target))
	assert.Equal(t, origin, target)
}

func TestCopyModelToProto(t *testing.T) {
	type OriginLocation struct {
		City      string
		Latitude  float64
		Longitude float64
	}
	type originModel struct {
		Name      string
		Age       int
		Money     float64
		BirthDay  time.Time
		IsDeleted bool
		Location  OriginLocation
	}
	type TargetLocation struct {
		City      string
		Latitude  float32
		Longitude float32
	}
	type targetModel struct {
		Id       string
		Name     string
		Age      int
		Money    float64
		BirthDay string
		Location TargetLocation
	}

	var target targetModel
	originTime := time.Now()
	origin := originModel{
		Name:      "MockModel",
		Age:       21,
		Money:     7.7,
		BirthDay:  originTime,
		IsDeleted: false,
		Location: OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
	}

	assert.Nil(t, Instance(nil).Install(RFC3339Convertor).From(origin).CopyTo(&target))
	assert.Equal(t, target, targetModel{
		Name:     "MockModel",
		Age:      21,
		Money:    7.7,
		BirthDay: originTime.Format(RFC3339Mili),
		Location: TargetLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
	})
}

func TestCopySlice(t *testing.T) {
	type OriginLocation struct {
		City      string
		Latitude  float64
		Longitude float64
	}
	type originModel struct {
		Name      string
		Age       int
		Money     float64
		BirthDay  time.Time
		IsDeleted bool
		Location  OriginLocation
		Detail    *OriginLocation
	}

	var targets []*originModel
	origin := originModel{
		Name:      "MockModel",
		Age:       21,
		Money:     7.7,
		BirthDay:  time.Now(),
		IsDeleted: false,
		Location: OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
		Detail: &OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
	}
	origins := []originModel{origin, origin}

	assert.Nil(t, Instance(nil).From(origins).CopyTo(&targets))
	assert.Equal(t, 2, len(targets))
	assert.Equal(t, origin, *targets[0])
	assert.Equal(t, origin, *targets[1])
}

func TestSkipExists(t *testing.T) {
	type OriginLocation struct {
		City      string
		Latitude  float64
		Longitude float64
	}
	type originModel struct {
		Name      string
		Age       int
		Money     float64
		BirthDay  time.Time
		IsDeleted bool
		Location  OriginLocation
		Detail    *OriginLocation
	}

	target := originModel{}
	origin := originModel{
		Name:      "MockModel",
		Age:       21,
		Money:     7.7,
		BirthDay:  time.Now(),
		IsDeleted: false,
		Location: OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
		Detail: &OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
	}

	target.Name = "I already have a name"
	assert.Nil(t, Instance(NewOption().SetOverwrite(false)).From(origin).CopyTo(&target))
	origin.Name = target.Name
	assert.Equal(t, origin, target)
}

func TestTransformerModelToProto(t *testing.T) {
	type Location struct {
		City      string
		Latitude  float64
		Longitude float64
	}
	type originModel struct {
		Name     string
		BirthDay time.Time
		StoreId  string
	}

	type targetModel struct {
		Id         string
		TargetName string
		Name       string
		CreatedAt  string
		Location   *Location
	}

	var targets []targetModel
	locationMapper := map[string]*Location{
		"12306": &Location{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
	}

	origins := []originModel{
		{
			Name:     "MockModel1",
			BirthDay: time.Now(),
			StoreId:  "12306",
		},
		{
			Name:     "MockModel2",
			BirthDay: time.Now(),
			StoreId:  "12345",
		},
	}
	assert.Nil(t, Instance(nil).
		RegisterTransformer(map[string]interface{}{
			"Location": func(storeId string) *Location {
				if location, ok := locationMapper[storeId]; ok {
					return location
				}
				return nil
			},
		}).
		RegisterResetDiffField([]DiffFieldPair{
			{Origin: "Name", Targets: []string{"TargetName", "Name"}},
			{Origin: "BirthDay", Targets: []string{"CreatedAt"}},
			{Origin: "StoreId", Targets: []string{"Location"}}},
		).Install(RFC3339Convertor).From(origins).CopyTo(&targets))

	assert.Equal(t, targetModel{
		TargetName: "MockModel1",
		Name:       "MockModel1",
		CreatedAt:  origins[0].BirthDay.Format(RFC3339Mili),
		Location:   locationMapper["12306"],
	}, targets[0])
	assert.Equal(t, targetModel{
		TargetName: "MockModel2",
		Name:       "MockModel2",
		CreatedAt:  origins[1].BirthDay.Format(RFC3339Mili),
		Location:   nil,
	}, targets[1])
}

func TestCopyModelToProtoWithMultiLevelAndTransformer(t *testing.T) {
	type Age struct {
		Value int
	}
	type OriginCityInfo struct {
		Age  Age
		Area float64
	}
	type TargetCityInfo struct {
		Age  Age
		Name string
	}
	type OriginLocation struct {
		City     string
		CityInfo OriginCityInfo
	}
	type originModel struct {
		Name     string
		Location OriginLocation
	}
	type TargetLocation struct {
		City         string
		CityName     string
		CityNickName string
		CityInfo     TargetCityInfo
	}
	type targetModel struct {
		Name string
		Loc  *TargetLocation
	}

	var targets []targetModel
	origins := []originModel{
		{
			Name: "MockModel",
			Location: OriginLocation{
				City: "ShangHai",
				CityInfo: OriginCityInfo{
					Age:  Age{Value: 1},
					Area: 1,
				},
			},
		},
	}
	assert.Nil(t, Instance(nil).RegisterTransformer(map[string]interface{}{
		"Loc.CityNickName": func(city string) string {
			return "Transformer city nick name"
		},
		"Loc.CityInfo.Age": func(city Age) Age {
			city.Value++
			return city
		},
	}).RegisterResetDiffField([]DiffFieldPair{
		{Origin: "Location", Targets: []string{"Loc"}},
		{Origin: "Location.City", Targets: []string{"Loc.CityName", "Loc.CityNickName"}},
		{Origin: "Location.CityInfo.Age", Targets: []string{"Loc.CityInfo.Age"}},
	}).From(origins).CopyTo(&targets))

	assert.Equal(t, targetModel{
		Name: "MockModel",
		Loc: &TargetLocation{
			City:         "ShangHai",
			CityName:     "ShangHai",
			CityNickName: "Transformer city nick name",
			CityInfo: TargetCityInfo{
				Age: Age{Value: 2},
			},
		},
	}, targets[0])
}

func TestCopyModelToProtoFromDeepLevel(t *testing.T) {
	type rChannel struct {
		Id     string
		Origin string
	}
	type modelChannel struct {
		Id     string
		Origin string
	}
	type ScoreInfo struct {
		Channel modelChannel
	}
	type response struct {
		Channel *rChannel
	}
	type model struct {
		Info ScoreInfo
	}

	resp := []response{}
	origin := []model{
		{
			Info: ScoreInfo{
				Channel: modelChannel{
					Id:     "123456",
					Origin: "retail",
				},
			},
		},
	}
	assert.Nil(t, Instance(nil).RegisterTransformer(Transformer{
		"Channel": func(s ScoreInfo) *rChannel {
			return &rChannel{
				Id:     s.Channel.Id,
				Origin: s.Channel.Origin,
			}
		},
	}).RegisterResetDiffField([]DiffFieldPair{
		{Origin: "Info", Targets: []string{"Channel"}},
	}).From(origin).CopyTo(&resp))

	assert.Equal(t, response{
		Channel: &rChannel{
			Id:     "123456",
			Origin: "retail",
		},
	}, resp[0])
}

func TestCopyModelToProtoWithTransformerMultiField(t *testing.T) {
	type originModel struct {
		Name     string
		Age      int
		BirthDay time.Time
	}
	type targetModel struct {
		Name    string
		NameArr []string
		Content *string
	}

	var targets []targetModel
	origins := []originModel{
		{
			Name:     "MockModel",
			BirthDay: time.Now(),
			Age:      18,
		},
	}
	assert.Nil(t, Instance(nil).RegisterResetDiffField([]DiffFieldPair{
		{Origin: "BirthDay", Targets: []string{"Name", "NameArr"}},
		{Origin: "Name", Targets: []string{"Name", "NameArr"}},
		{Origin: "Age", Targets: []string{"Content"}},
	}).RegisterTransformer(map[string]interface{}{
		"NameArr": func(value interface{}, originFieldKey string, target []string) []string {
			switch originFieldKey {
			case "BirthDay":
				target = append(target, value.(time.Time).Format(time.Kitchen))
			case "Name":
				target = append(target, value.(string))
			}
			return target
		},
		"Content": func(age int, originFieldKey string, target *string) *string {
			switch originFieldKey {
			case "Age":
				result := strconv.FormatInt(int64(age), 10)
				return &result
			}
			return nil
		},
		"Name": func(value interface{}, originFieldKey string, target string, targetKey string) string {
			switch originFieldKey {
			case "BirthDay":
				target = target + ", I was born on " + value.(time.Time).Format(time.Kitchen) + "."
			case "Name":
				target = "My name is " + value.(string)
			}
			return target
		},
	}).From(origins).CopyTo(&targets))

	age := "18"
	t.Log()
	assert.Equal(t, targetModel{
		Name:    "My name is MockModel, I was born on " + time.Now().Format(time.Kitchen) + ".",
		NameArr: []string{"MockModel", time.Now().Format(time.Kitchen)},
		Content: &age,
	}, targets[0])
}

func TestCopyModelToModelWithIgnoreInvalidOption(t *testing.T) {
	type originModel struct {
		BirthDay  time.Time
		IsDeleted bool
	}
	type targetModel struct {
		Id        int
		BirthDay  int
		IsDeleted string
	}

	target := targetModel{}
	origin := originModel{
		BirthDay:  time.Now(),
		IsDeleted: false,
	}

	assert.Nil(t, Instance(NewOption().SetIgnoreEmpty(true).SetOverwrite(false)).RegisterIgnoreTargetFields([]FieldKey{"Id", "BirthDay"}).From(origin).CopyTo(&target))
}

func TestCopyModelToProtoModelWithOverwriteOriginalCopyFieldOption(t *testing.T) {
	type originModel struct {
		ProductId string
	}
	type targetModel struct {
		Id        string
		ProductId string
	}

	target := targetModel{}
	origin := originModel{
		ProductId: "TestProductId",
	}
	copier := Instance(NewOption().SetOverwriteOriginalCopyField(true))
	assert.Nil(t, copier.RegisterResetDiffField([]DiffFieldPair{{Origin: "Id", Targets: []string{"ProductId"}}}).From(origin).CopyTo(&target))
	assert.Equal(t, targetModel{
		Id: "",
	}, target)

	targets := []targetModel{}
	origins := []originModel{
		{ProductId: "TestProductId1"},
		{ProductId: "TestProductId2"},
	}
	assert.Nil(t, copier.RegisterResetDiffField([]DiffFieldPair{{Origin: "Id", Targets: []string{"Id", "ProductId"}}}).From(origins).CopyTo(&targets))
}

func TestDoubleModelIntoOneProtoModel(t *testing.T) {
	type Template struct {
		Id        string `bson:"id"`
		AppSecret bool   `bson:"appSecret"`
		Title     string `bson:"title"`
	}
	type Task struct {
		Name     string   `bson:"name"`
		Type     string   `bson:"type"`
		Template Template `bson:"template"`
	}
	type TaskDetail struct {
		Id        string
		Name      string
		Type      string
		AppSecret bool
		Title     string
	}

	task := Task{
		Name: "task",
		Type: "task",
		Template: Template{
			Id:        "templateId",
			AppSecret: true,
			Title:     "templateTitle",
		},
	}
	taskDetail := TaskDetail{}
	err := Instance(nil).RegisterResetDiffField([]DiffFieldPair{
		{
			Origin:  "Template.Id",
			Targets: []string{"Id"},
		},
		{
			Origin:  "Template.Title",
			Targets: []string{"Title"},
		},
		{
			Origin:  "Template.AppSecret",
			Targets: []string{"AppSecret"},
		},
	}).From(task).CopyTo(&taskDetail)
	assert.Nil(t, err)
	assert.Equal(t, task.Template.Id, taskDetail.Id)
	assert.Equal(t, task.Template.Title, taskDetail.Title)
	assert.Equal(t, task.Template.AppSecret, taskDetail.AppSecret)
}

func TestCopierUnexportedField(t *testing.T) {
	type incScoreOption struct {
		Name string
	}
	type ScoreInfo struct {
		Score          int
		Description    *string
		businessId     string
		reason         *string
		incScoreOption *incScoreOption
	}

	scoreDescription := "inc by register"
	reason := "task"
	scoreInfo := &ScoreInfo{
		Score:       12,
		Description: &scoreDescription,
		businessId:  "2022-03-09",
		reason:      &reason,
		incScoreOption: &incScoreOption{
			Name: "origin",
		},
	}

	result := ScoreInfo{}
	err := Instance(NewOption().SetCopyUnexported(true)).From(scoreInfo).CopyTo(&result)

	assert.Nil(t, err)
	assert.Equal(t, result.incScoreOption.Name, scoreInfo.incScoreOption.Name)
	assert.Equal(t, result.Score, scoreInfo.Score)
	assert.Equal(t, *result.Description, *scoreInfo.Description)
	assert.Equal(t, result.Description, scoreInfo.Description)
	assert.Equal(t, result.reason, scoreInfo.reason)
	assert.NotSame(t, result.Description, scoreInfo.Description)
	assert.NotSame(t, result.reason, scoreInfo.reason)
	assert.Equal(t, result.businessId, scoreInfo.businessId)
}

func TestCopierWithIgnoreDeepEmpty(t *testing.T) {
	type Property1 struct {
		Name string
		Jobs []string
	}
	type Info1 struct {
		Property *Property1
		InfoName string
	}
	type Tag1 struct {
		TagName string
	}
	type EcProduct1 struct {
		Info *Info1
		Tags *[]Tag1
	}

	from := &EcProduct1{
		Info: &Info1{
			Property: &Property1{},
			InfoName: "info",
		},
		Tags: &[]Tag1{},
	}

	type Property2 struct {
		Name string
	}
	type Info2 struct {
		Property *Property2
		InfoName string
	}
	type Tag2 struct {
		TagName string
	}
	type EcProduct2 struct {
		Info *Info2
		Tags *[]Tag2
	}

	to := &EcProduct2{}
	err := Instance(NewOption().SetIgnoreDeepEmpty(true)).From(from).CopyTo(to)

	assert.Nil(t, err)
	assert.NotNil(t, to.Info)
	assert.Nil(t, to.Info.Property)
	assert.Nil(t, to.Tags)

	to1 := &EcProduct1{}
	err = Instance(NewOption().SetIgnoreDeepEmpty(true)).From(from).CopyTo(to1)

	assert.Nil(t, err)
	assert.NotNil(t, to1.Info)
	assert.Nil(t, to1.Info.Property)

	err = Instance(NewOption().SetIgnoreDeepEmpty(false)).From(from).CopyTo(to)

	assert.Nil(t, err)
	assert.Equal(t, to.Info.Property.Name, "")
	assert.NotNil(t, to.Tags)
}

func TestIgnoreDeepEmpty(t *testing.T) {
	type MemberLimit struct {
		Type            string
		Operator        string
		Tags            []string
		StaticGroupIds  []string
		DynamicGroupIds []string
	}
	type PurchaseLimit struct {
		// 限制指定人群
		Member     *MemberLimit
		Count      int64
		PeriodType string
	}
	type UpdateECProduct struct {
		PurchaseLimit *PurchaseLimit
	}
	type ECProduct struct {
		PurchaseLimit *PurchaseLimit
	}
	type ProductResponse struct {
		Ec *ECProduct
	}
	ec := &UpdateECProduct{}
	productDetail := &ProductResponse{
		Ec: &ECProduct{
			PurchaseLimit: &PurchaseLimit{
				Member: &MemberLimit{
					Type:            "groups",
					Operator:        "IN",
					Tags:            nil,
					StaticGroupIds:  nil,
					DynamicGroupIds: []string{"639ae585dc5fd52014512ac6"},
				},
				Count:      11,
				PeriodType: "13",
			},
		},
	}
	Instance(NewOption().SetIgnoreDeepEmpty(true)).From(productDetail.Ec).CopyTo(ec)
	assert.Equal(t, productDetail.Ec.PurchaseLimit.Member.Type, ec.PurchaseLimit.Member.Type)
}

func TestIgnoreEmptyFields(t *testing.T) {
	type EventRollbacker struct {
		EventId string
	}
	type ProtoEventRollbacker struct {
		EventId string
	}
	type MemberTask struct {
		Type            string
		IsEnabled       bool
		Count           int
		EventRollbacker []EventRollbacker
	}

	type ProtoMemberTask struct {
		Type            string
		IsEnabled       bool
		EventRollbacker *[]ProtoEventRollbacker
	}
	task := &MemberTask{
		Count:     1,
		IsEnabled: true,
	}
	protoTask := &ProtoMemberTask{
		Type: "proto",
	}

	task.EventRollbacker = []EventRollbacker{
		{"order"},
	}

	Instance(NewOption()).From(protoTask).CopyTo(task)
	assert.Equal(t, len(task.EventRollbacker), 0)
	assert.Equal(t, task.Type, "proto")

	task.EventRollbacker = []EventRollbacker{
		{"order"},
	}

	Instance(NewOption().SetIgnoreEmptyField([]string{"EventRollbacker"})).From(protoTask).CopyTo(task)
	assert.Equal(t, len(task.EventRollbacker), 1)
	assert.Equal(t, task.EventRollbacker[0].EventId, "order")
	assert.Equal(t, task.Type, "proto")
	assert.Equal(t, task.Count, 1)
}

func TestCopierMirror(t *testing.T) {
	type EventRollbacker struct {
		EventId string
	}
	type ProtoEventRollbacker struct {
		EventId string
	}
	type MemberTask struct {
		Type            string
		IsEnabled       bool
		Count           int
		EventRollbacker []EventRollbacker
	}

	type ProtoMemberTask struct {
		Type            string
		IsEnabled       bool
		EventRollbacker *[]ProtoEventRollbacker
	}
	task := &MemberTask{
		Count:     1,
		IsEnabled: true,
	}
	protoTask := &ProtoMemberTask{
		Type: "proto",
	}

	task.EventRollbacker = []EventRollbacker{
		{"order"},
	}

	task = InstanceMirror[*MemberTask](NewOption()).Mirror(protoTask)
	assert.Equal(t, len(task.EventRollbacker), 0)
	assert.Equal(t, task.Type, "proto")
}
