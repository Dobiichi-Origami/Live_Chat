// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package http

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	entities "liveChat/entities"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonDe1d482eDecodeLiveChat(in *jlexer.Lexer, out *UserInfoBody) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.Id = int64(in.Int64())
		case "username":
			out.Username = string(in.String())
		case "avatar":
			out.UserAvatar = string(in.String())
		case "introduction":
			out.UserIntroduction = string(in.String())
		case "friendships":
			if in.IsNull() {
				in.Skip()
				out.Friendships = nil
			} else {
				in.Delim('[')
				if out.Friendships == nil {
					if !in.IsDelim(']') {
						out.Friendships = make([]entities.Friendship, 0, 0)
					} else {
						out.Friendships = []entities.Friendship{}
					}
				} else {
					out.Friendships = (out.Friendships)[:0]
				}
				for !in.IsDelim(']') {
					var v1 entities.Friendship
					easyjsonDe1d482eDecodeLiveChatEntities(in, &v1)
					out.Friendships = append(out.Friendships, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "groupList":
			if in.IsNull() {
				in.Skip()
				out.Groups = nil
			} else {
				in.Delim('[')
				if out.Groups == nil {
					if !in.IsDelim(']') {
						out.Groups = make([]entities.GroupMember, 0, 0)
					} else {
						out.Groups = []entities.GroupMember{}
					}
				} else {
					out.Groups = (out.Groups)[:0]
				}
				for !in.IsDelim(']') {
					var v2 entities.GroupMember
					easyjsonDe1d482eDecodeLiveChatEntities1(in, &v2)
					out.Groups = append(out.Groups, v2)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "status":
			out.Status = int32(in.Int32())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonDe1d482eEncodeLiveChat(out *jwriter.Writer, in UserInfoBody) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.Int64(int64(in.Id))
	}
	{
		const prefix string = ",\"username\":"
		out.RawString(prefix)
		out.String(string(in.Username))
	}
	{
		const prefix string = ",\"avatar\":"
		out.RawString(prefix)
		out.String(string(in.UserAvatar))
	}
	{
		const prefix string = ",\"introduction\":"
		out.RawString(prefix)
		out.String(string(in.UserIntroduction))
	}
	{
		const prefix string = ",\"friendships\":"
		out.RawString(prefix)
		if in.Friendships == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v3, v4 := range in.Friendships {
				if v3 > 0 {
					out.RawByte(',')
				}
				easyjsonDe1d482eEncodeLiveChatEntities(out, v4)
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"groupList\":"
		out.RawString(prefix)
		if in.Groups == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v5, v6 := range in.Groups {
				if v5 > 0 {
					out.RawByte(',')
				}
				easyjsonDe1d482eEncodeLiveChatEntities1(out, v6)
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"status\":"
		out.RawString(prefix)
		out.Int32(int32(in.Status))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v UserInfoBody) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDe1d482eEncodeLiveChat(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v UserInfoBody) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDe1d482eEncodeLiveChat(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *UserInfoBody) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDe1d482eDecodeLiveChat(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *UserInfoBody) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDe1d482eDecodeLiveChat(l, v)
}
func easyjsonDe1d482eDecodeLiveChatEntities1(in *jlexer.Lexer, out *entities.GroupMember) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "groupId":
			out.GroupId = int64(in.Int64())
		case "memberId":
			out.MemberId = int64(in.Int64())
		case "isAdministrator":
			out.IsAdministrator = bool(in.Bool())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonDe1d482eEncodeLiveChatEntities1(out *jwriter.Writer, in entities.GroupMember) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"groupId\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.GroupId))
	}
	{
		const prefix string = ",\"memberId\":"
		out.RawString(prefix)
		out.Int64(int64(in.MemberId))
	}
	{
		const prefix string = ",\"isAdministrator\":"
		out.RawString(prefix)
		out.Bool(bool(in.IsAdministrator))
	}
	out.RawByte('}')
}
func easyjsonDe1d482eDecodeLiveChatEntities(in *jlexer.Lexer, out *entities.Friendship) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "selfId":
			out.SelfId = int64(in.Int64())
		case "friendId":
			out.FriendId = int64(in.Int64())
		case "chatId":
			out.ChatId = int64(in.Int64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonDe1d482eEncodeLiveChatEntities(out *jwriter.Writer, in entities.Friendship) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"selfId\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.SelfId))
	}
	{
		const prefix string = ",\"friendId\":"
		out.RawString(prefix)
		out.Int64(int64(in.FriendId))
	}
	{
		const prefix string = ",\"chatId\":"
		out.RawString(prefix)
		out.Int64(int64(in.ChatId))
	}
	out.RawByte('}')
}
func easyjsonDe1d482eDecodeLiveChat1(in *jlexer.Lexer, out *SuccessBody) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "status":
			out.Status = int32(in.Int32())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonDe1d482eEncodeLiveChat1(out *jwriter.Writer, in SuccessBody) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"status\":"
		out.RawString(prefix[1:])
		out.Int32(int32(in.Status))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v SuccessBody) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDe1d482eEncodeLiveChat1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v SuccessBody) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDe1d482eEncodeLiveChat1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *SuccessBody) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDe1d482eDecodeLiveChat1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *SuccessBody) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDe1d482eDecodeLiveChat1(l, v)
}
func easyjsonDe1d482eDecodeLiveChat2(in *jlexer.Lexer, out *ResponseHeader) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "status":
			out.Status = int32(in.Int32())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonDe1d482eEncodeLiveChat2(out *jwriter.Writer, in ResponseHeader) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"status\":"
		out.RawString(prefix[1:])
		out.Int32(int32(in.Status))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ResponseHeader) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDe1d482eEncodeLiveChat2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ResponseHeader) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDe1d482eEncodeLiveChat2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ResponseHeader) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDe1d482eDecodeLiveChat2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ResponseHeader) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDe1d482eDecodeLiveChat2(l, v)
}
func easyjsonDe1d482eDecodeLiveChat3(in *jlexer.Lexer, out *RegisterOrLoginBody) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "token":
			out.Token = string(in.String())
		case "status":
			out.Status = int32(in.Int32())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonDe1d482eEncodeLiveChat3(out *jwriter.Writer, in RegisterOrLoginBody) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"token\":"
		out.RawString(prefix[1:])
		out.String(string(in.Token))
	}
	{
		const prefix string = ",\"status\":"
		out.RawString(prefix)
		out.Int32(int32(in.Status))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v RegisterOrLoginBody) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDe1d482eEncodeLiveChat3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v RegisterOrLoginBody) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDe1d482eEncodeLiveChat3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *RegisterOrLoginBody) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDe1d482eDecodeLiveChat3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *RegisterOrLoginBody) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDe1d482eDecodeLiveChat3(l, v)
}
func easyjsonDe1d482eDecodeLiveChat4(in *jlexer.Lexer, out *GroupInfoBody) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.Id = int64(in.Int64())
		case "ownerId":
			out.Owner = int64(in.Int64())
		case "name":
			out.Name = string(in.String())
		case "introduction":
			out.Introduction = string(in.String())
		case "avatar":
			out.Avatar = string(in.String())
		case "members":
			if in.IsNull() {
				in.Skip()
				out.Members = nil
			} else {
				in.Delim('[')
				if out.Members == nil {
					if !in.IsDelim(']') {
						out.Members = make([]entities.GroupMember, 0, 0)
					} else {
						out.Members = []entities.GroupMember{}
					}
				} else {
					out.Members = (out.Members)[:0]
				}
				for !in.IsDelim(']') {
					var v7 entities.GroupMember
					easyjsonDe1d482eDecodeLiveChatEntities1(in, &v7)
					out.Members = append(out.Members, v7)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "status":
			out.Status = int32(in.Int32())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonDe1d482eEncodeLiveChat4(out *jwriter.Writer, in GroupInfoBody) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.Int64(int64(in.Id))
	}
	{
		const prefix string = ",\"ownerId\":"
		out.RawString(prefix)
		out.Int64(int64(in.Owner))
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"introduction\":"
		out.RawString(prefix)
		out.String(string(in.Introduction))
	}
	{
		const prefix string = ",\"avatar\":"
		out.RawString(prefix)
		out.String(string(in.Avatar))
	}
	{
		const prefix string = ",\"members\":"
		out.RawString(prefix)
		if in.Members == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v8, v9 := range in.Members {
				if v8 > 0 {
					out.RawByte(',')
				}
				easyjsonDe1d482eEncodeLiveChatEntities1(out, v9)
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"status\":"
		out.RawString(prefix)
		out.Int32(int32(in.Status))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v GroupInfoBody) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDe1d482eEncodeLiveChat4(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v GroupInfoBody) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDe1d482eEncodeLiveChat4(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *GroupInfoBody) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDe1d482eDecodeLiveChat4(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *GroupInfoBody) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDe1d482eDecodeLiveChat4(l, v)
}
func easyjsonDe1d482eDecodeLiveChat5(in *jlexer.Lexer, out *FailBody) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "reason":
			out.Reason = string(in.String())
		case "status":
			out.Status = int32(in.Int32())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonDe1d482eEncodeLiveChat5(out *jwriter.Writer, in FailBody) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"reason\":"
		out.RawString(prefix[1:])
		out.String(string(in.Reason))
	}
	{
		const prefix string = ",\"status\":"
		out.RawString(prefix)
		out.Int32(int32(in.Status))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v FailBody) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDe1d482eEncodeLiveChat5(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v FailBody) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDe1d482eEncodeLiveChat5(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *FailBody) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDe1d482eDecodeLiveChat5(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *FailBody) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDe1d482eDecodeLiveChat5(l, v)
}
