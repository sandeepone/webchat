package controllers

import (
	"github.com/robfig/revel"
	"webchat/app/chatserver"
	"webchat/app/form"
	"webchat/app/model"
    //"log"
)

type Rooms struct {
	*Application
}

type RoomApi struct {
	*revel.Controller
}


func (c Rooms) Index(p int) revel.Result {
	if p == 0 { p = 1 }

	rooms := model.FindOnePage(p)

    // generate roomlist with recent users
    var roomLists []*model.RoomList
    for _, room := range rooms {
        recentUsers := room.GetRecentUsers() // get []*RecentUser
        //log.Println("recentUsers is:", recentUsers)

        rl := &model.RoomList{
            Room: &room,
            RecentUsers: recentUsers,
        }

        roomLists = append(roomLists, rl)
    }

	allPage := (model.RoomCount() + model.PageSize - 1) / model.PageSize

	return c.Render(p, allPage, roomLists)
}

func (c Rooms) New() revel.Result {

	if !isLogin(c.Controller) {
		c.Flash.Error("Please login first")
		return c.Redirect(Application.Index)
	}

	return c.Render()
}

func (c Rooms) Create(rf *form.RoomForm) revel.Result {

	if !isLogin(c.Controller) {
		c.Flash.Error("Please login first")
		return c.Redirect(Application.Index)
	}

	rf.UserId = CurrentUser(c.Controller).Id

	rf.Validate(c.Validation)

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Rooms.New)
	}
	room := model.NewRoom(rf)

	if _, err := room.Save(); err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect(Rooms.New)
	}

	// run activeroom
	activeroom := chatserver.NewActiveRoom(room.RoomKey)
	go activeroom.Run()
	ChatServer.ActiveRooms.PushBack(activeroom)

	return c.Redirect("/r/%s", room.RoomKey)
}

func (c Rooms) Show(roomkey string) revel.Result {
	if !isLogin(c.Controller) {
		c.Flash.Error("Please login first")
		return c.Redirect(Application.Index)
	}

    currentUser := CurrentUser(c.Controller)

	room := model.FindRoomByRoomKey(roomkey)
	activeRoom := ChatServer.GetActiveRoom(roomkey)
    activeRoom.AddUserToRecent(currentUser)

    // user list
	users := activeRoom.UserList()

    // room list 
    rooms := model.FindRoomByUserId(currentUser.Id)
    // user avatar
	userAvatar := currentUser.AvatarUrl()

	return c.Render(room, users, userAvatar, rooms)
}

func (c Rooms) Edit(roomkey string) revel.Result {

	if !isLogin(c.Controller) {
		c.Flash.Error("Please login first")
		return c.Redirect(Application.Index)
	}

	room := model.FindRoomByRoomKey(roomkey)

	return c.Render(room)
}

func (c Rooms) Update(roomkey string, updateroom *form.UpdateRoom) revel.Result {

	if !isLogin(c.Controller) {
		c.Flash.Error("Please login first")
		return c.Redirect(Application.Index)
	}

	room := model.FindRoomByRoomKey(roomkey)

	if err := room.Update(updateroom); err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect("/r/%s/edit", room.RoomKey)
	}

	c.Flash.Success("update success")
	return c.Redirect("/r/%s/edit", room.RoomKey)
}


type UserList struct {
	Users []*chatserver.UserInfo
}

func (c RoomApi) Users(roomkey string) revel.Result {

	// get a activeRoom and get room's user list 
	activeroom := ChatServer.GetActiveRoom(roomkey)
	users := activeroom.UserList()

	userList := &UserList{
		Users: users,
	}

	return c.RenderJson(userList)
}
