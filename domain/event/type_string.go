// Code generated by "stringer -type Type"; DO NOT EDIT.

package event

import "strconv"

const _Type_name = "TypeNoneTypeErrorRaisedTypeUserCreatedTypeUserDeletedTypeUserAddedFriendTypeRoomCreatedTypeRoomDeletedTypeRoomAddedMemberTypeRoomRemoveMemberTypeRoomMessagesReadByUserTypeMessageCreatedTypeActiveClientActivatedTypeActiveClientInactivated"

var _Type_index = [...]uint8{0, 8, 23, 38, 53, 72, 87, 102, 121, 141, 167, 185, 210, 237}

func (i Type) String() string {
	if i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
