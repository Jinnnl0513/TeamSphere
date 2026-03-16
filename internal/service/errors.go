package service

import "errors"

var (
	ErrNoPermission = errors.New("no permission")
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrWrongPassword = errors.New("wrong password")
	ErrOwnsRooms     = errors.New("account owns rooms")
)

var (
	ErrRoomNotFound      = errors.New("room not found")
	ErrNotRoomMember     = errors.New("not a room member")
	ErrAlreadyMember     = errors.New("already a room member")
	ErrTargetNotMember   = errors.New("target user is not a room member")
	ErrRoomNameTaken     = errors.New("you already have a room with this name")
	ErrCannotLeaveOwner  = errors.New("owner cannot leave without transferring ownership")
	ErrCannotActOnHigher = errors.New("cannot act on member with equal or higher role")
	ErrInviteNotFriend   = errors.New("invitee is not your friend")
	ErrInviteNotFound    = errors.New("invite not found")
	ErrInvitePending     = errors.New("a pending invite already exists")
	ErrJoinApprovalRequired = errors.New("join requires approval")
	ErrJoinInviteOnly       = errors.New("room is invite only")
	ErrJoinRequestPending   = errors.New("join request already pending")
	ErrRoomReadOnly         = errors.New("room is read only")
	ErrRoomSendForbidden    = errors.New("send forbidden by room policy")
	ErrRoomUploadForbidden  = errors.New("upload forbidden by room policy")
	ErrContentBlocked       = errors.New("content blocked by policy")
)

var (
	ErrMessageNotFound = errors.New("message not found")
	ErrRecallTimeout   = errors.New("recall time window exceeded")
	ErrRecallForbidden = errors.New("no permission to recall this message")
	ErrAlreadyRecalled = errors.New("message already recalled")
	ErrEditForbidden   = errors.New("only the sender can edit this message")
	ErrInvalidParams   = errors.New("invalid params")
)

var (
	ErrFriendSelf        = errors.New("cannot add yourself as friend")
	ErrFriendExists      = errors.New("friendship already exists")
	ErrFriendReqPending  = errors.New("a pending request already exists")
	ErrFriendReqNotFound = errors.New("friend request not found")
	ErrNotFriends        = errors.New("not friends")
)

var (
	ErrFileTooLarge       = errors.New("file too large")
	ErrFileTypeNotAllowed = errors.New("file type not allowed")
)

var (
	ErrAdminSelfRole     = errors.New("cannot change your own role")
	ErrAdminSelfDelete   = errors.New("cannot delete your own account via admin")
	ErrAdminInvalidRole  = errors.New("invalid role: must be user or admin")
	ErrAdminDeleteOwner  = errors.New("cannot delete the system owner")
	ErrAdminOnlyOneOwner = errors.New("there can only be one system owner")
)
