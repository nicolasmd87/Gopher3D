// Code generated by cmd/codegen from https://github.com/AllenDang/cimgui-go.
// DO NOT EDIT.

package imgui

// #include <stdlib.h>
// #include <memory.h>
// #include "extra_types.h"
// #include "cimmarkdown_wrapper.h"
import "C"
import "unsafe"

type Emphasis struct {
	FieldState EmphasisState
	FieldText  TextBlock
	FieldSym   rune
}

func (self Emphasis) handle() (result *C.Emphasis, releaseFn func()) {
	result = new(C.Emphasis)
	FieldState := self.FieldState

	result.state = C.EmphasisState(FieldState)
	FieldText := self.FieldText
	FieldTextArg, FieldTextFin := FieldText.c()
	result.text = FieldTextArg
	FieldSym := self.FieldSym

	result.sym = C.char(FieldSym)
	releaseFn = func() {
		FieldTextFin()
	}
	return result, releaseFn
}

func (self Emphasis) c() (result C.Emphasis, fin func()) {
	resultPtr, finFn := self.handle()
	return *resultPtr, finFn
}

func newEmphasisFromC(cvalue *C.Emphasis) *Emphasis {
	result := new(Emphasis)
	result.FieldState = EmphasisState(cvalue.state)
	result.FieldText = *newTextBlockFromC(func() *C.TextBlock { result := cvalue.text; return &result }())

	result.FieldSym = rune(cvalue.sym)
	return result
}

type Line struct {
	FieldIsHeading            bool
	FieldIsEmphasis           bool
	FieldIsUnorderedListStart bool
	FieldIsLeadingSpace       bool
	FieldLeadSpaceCount       int32
	FieldHeadingCount         int32
	FieldEmphasisCount        int32
	FieldLineStart            int32
	FieldLineEnd              int32
	FieldLastRenderPosition   int32
}

func (self Line) handle() (result *C.Line, releaseFn func()) {
	result = new(C.Line)
	FieldIsHeading := self.FieldIsHeading

	result.isHeading = C.bool(FieldIsHeading)
	FieldIsEmphasis := self.FieldIsEmphasis

	result.isEmphasis = C.bool(FieldIsEmphasis)
	FieldIsUnorderedListStart := self.FieldIsUnorderedListStart

	result.isUnorderedListStart = C.bool(FieldIsUnorderedListStart)
	FieldIsLeadingSpace := self.FieldIsLeadingSpace

	result.isLeadingSpace = C.bool(FieldIsLeadingSpace)
	FieldLeadSpaceCount := self.FieldLeadSpaceCount

	result.leadSpaceCount = C.int(FieldLeadSpaceCount)
	FieldHeadingCount := self.FieldHeadingCount

	result.headingCount = C.int(FieldHeadingCount)
	FieldEmphasisCount := self.FieldEmphasisCount

	result.emphasisCount = C.int(FieldEmphasisCount)
	FieldLineStart := self.FieldLineStart

	result.lineStart = C.int(FieldLineStart)
	FieldLineEnd := self.FieldLineEnd

	result.lineEnd = C.int(FieldLineEnd)
	FieldLastRenderPosition := self.FieldLastRenderPosition

	result.lastRenderPosition = C.int(FieldLastRenderPosition)
	releaseFn = func() {
	}
	return result, releaseFn
}

func (self Line) c() (result C.Line, fin func()) {
	resultPtr, finFn := self.handle()
	return *resultPtr, finFn
}

func newLineFromC(cvalue *C.Line) *Line {
	result := new(Line)
	result.FieldIsHeading = cvalue.isHeading == C.bool(true)
	result.FieldIsEmphasis = cvalue.isEmphasis == C.bool(true)
	result.FieldIsUnorderedListStart = cvalue.isUnorderedListStart == C.bool(true)
	result.FieldIsLeadingSpace = cvalue.isLeadingSpace == C.bool(true)
	result.FieldLeadSpaceCount = int32(cvalue.leadSpaceCount)
	result.FieldHeadingCount = int32(cvalue.headingCount)
	result.FieldEmphasisCount = int32(cvalue.emphasisCount)
	result.FieldLineStart = int32(cvalue.lineStart)
	result.FieldLineEnd = int32(cvalue.lineEnd)
	result.FieldLastRenderPosition = int32(cvalue.lastRenderPosition)
	return result
}

type Link struct {
	FieldState             LinkState
	FieldText              TextBlock
	FieldUrl               TextBlock
	FieldIsImage           bool
	FieldNum_brackets_open int32
}

func (self Link) handle() (result *C.Link, releaseFn func()) {
	result = new(C.Link)
	FieldState := self.FieldState

	result.state = C.LinkState(FieldState)
	FieldText := self.FieldText
	FieldTextArg, FieldTextFin := FieldText.c()
	result.text = FieldTextArg
	FieldUrl := self.FieldUrl
	FieldUrlArg, FieldUrlFin := FieldUrl.c()
	result.url = FieldUrlArg
	FieldIsImage := self.FieldIsImage

	result.isImage = C.bool(FieldIsImage)
	FieldNum_brackets_open := self.FieldNum_brackets_open

	result.num_brackets_open = C.int(FieldNum_brackets_open)
	releaseFn = func() {
		FieldTextFin()
		FieldUrlFin()
	}
	return result, releaseFn
}

func (self Link) c() (result C.Link, fin func()) {
	resultPtr, finFn := self.handle()
	return *resultPtr, finFn
}

func newLinkFromC(cvalue *C.Link) *Link {
	result := new(Link)
	result.FieldState = LinkState(cvalue.state)
	result.FieldText = *newTextBlockFromC(func() *C.TextBlock { result := cvalue.text; return &result }())

	result.FieldUrl = *newTextBlockFromC(func() *C.TextBlock { result := cvalue.url; return &result }())

	result.FieldIsImage = cvalue.isImage == C.bool(true)
	result.FieldNum_brackets_open = int32(cvalue.num_brackets_open)
	return result
}

type MarkdownConfig struct {
	// TODO: contains unsupported fields
	data unsafe.Pointer
}

func (self MarkdownConfig) handle() (result *C.MarkdownConfig, releaseFn func()) {
	result = (*C.MarkdownConfig)(self.data)
	return result, func() {}
}

func (self MarkdownConfig) c() (result C.MarkdownConfig, fin func()) {
	resultPtr, finFn := self.handle()
	return *resultPtr, finFn
}

func newMarkdownConfigFromC(cvalue *C.MarkdownConfig) *MarkdownConfig {
	result := new(MarkdownConfig)
	result.data = unsafe.Pointer(cvalue)
	return result
}

type MarkdownFormatInfo struct {
	// TODO: contains unsupported fields
	data unsafe.Pointer
}

func (self MarkdownFormatInfo) handle() (result *C.MarkdownFormatInfo, releaseFn func()) {
	result = (*C.MarkdownFormatInfo)(self.data)
	return result, func() {}
}

func (self MarkdownFormatInfo) c() (result C.MarkdownFormatInfo, fin func()) {
	resultPtr, finFn := self.handle()
	return *resultPtr, finFn
}

func newMarkdownFormatInfoFromC(cvalue *C.MarkdownFormatInfo) *MarkdownFormatInfo {
	result := new(MarkdownFormatInfo)
	result.data = unsafe.Pointer(cvalue)
	return result
}

type MarkdownHeadingFormat struct {
	FieldFont      *Font
	FieldSeparator bool
}

func (self MarkdownHeadingFormat) handle() (result *C.MarkdownHeadingFormat, releaseFn func()) {
	result = new(C.MarkdownHeadingFormat)
	FieldFont := self.FieldFont
	FieldFontArg, FieldFontFin := FieldFont.handle()
	result.font = FieldFontArg
	FieldSeparator := self.FieldSeparator

	result.separator = C.bool(FieldSeparator)
	releaseFn = func() {
		FieldFontFin()
	}
	return result, releaseFn
}

func (self MarkdownHeadingFormat) c() (result C.MarkdownHeadingFormat, fin func()) {
	resultPtr, finFn := self.handle()
	return *resultPtr, finFn
}

func newMarkdownHeadingFormatFromC(cvalue *C.MarkdownHeadingFormat) *MarkdownHeadingFormat {
	result := new(MarkdownHeadingFormat)
	result.FieldFont = newFontFromC(cvalue.font)
	result.FieldSeparator = cvalue.separator == C.bool(true)
	return result
}

type MarkdownImageData struct {
	FieldIsValid         bool
	FieldUseLinkCallback bool
	FieldUser_texture_id TextureID
	FieldSize            Vec2
	FieldUv0             Vec2
	FieldUv1             Vec2
	FieldTint_col        Vec4
	FieldBorder_col      Vec4
}

func (self MarkdownImageData) handle() (result *C.MarkdownImageData, releaseFn func()) {
	result = new(C.MarkdownImageData)
	FieldIsValid := self.FieldIsValid

	result.isValid = C.bool(FieldIsValid)
	FieldUseLinkCallback := self.FieldUseLinkCallback

	result.useLinkCallback = C.bool(FieldUseLinkCallback)
	FieldUser_texture_id := self.FieldUser_texture_id

	result.user_texture_id = C.ImTextureID(FieldUser_texture_id)
	FieldSize := self.FieldSize

	result.size = FieldSize.toC()
	FieldUv0 := self.FieldUv0

	result.uv0 = FieldUv0.toC()
	FieldUv1 := self.FieldUv1

	result.uv1 = FieldUv1.toC()
	FieldTint_col := self.FieldTint_col

	result.tint_col = FieldTint_col.toC()
	FieldBorder_col := self.FieldBorder_col

	result.border_col = FieldBorder_col.toC()
	releaseFn = func() {
	}
	return result, releaseFn
}

func (self MarkdownImageData) c() (result C.MarkdownImageData, fin func()) {
	resultPtr, finFn := self.handle()
	return *resultPtr, finFn
}

func newMarkdownImageDataFromC(cvalue *C.MarkdownImageData) *MarkdownImageData {
	result := new(MarkdownImageData)
	result.FieldIsValid = cvalue.isValid == C.bool(true)
	result.FieldUseLinkCallback = cvalue.useLinkCallback == C.bool(true)
	result.FieldUser_texture_id = TextureID(cvalue.user_texture_id)
	result.FieldSize = *(&Vec2{}).fromC(cvalue.size)
	result.FieldUv0 = *(&Vec2{}).fromC(cvalue.uv0)
	result.FieldUv1 = *(&Vec2{}).fromC(cvalue.uv1)
	result.FieldTint_col = *(&Vec4{}).fromC(cvalue.tint_col)
	result.FieldBorder_col = *(&Vec4{}).fromC(cvalue.border_col)
	return result
}

type MarkdownLinkCallbackData struct {
	FieldText       string
	FieldTextLength int32
	FieldLink       string
	FieldLinkLength int32
	FieldUserData   unsafe.Pointer
	FieldIsImage    bool
}

func (self MarkdownLinkCallbackData) handle() (result *C.MarkdownLinkCallbackData, releaseFn func()) {
	result = new(C.MarkdownLinkCallbackData)
	FieldText := self.FieldText
	FieldTextArg, FieldTextFin := WrapString(FieldText)
	result.text = FieldTextArg
	FieldTextLength := self.FieldTextLength

	result.textLength = C.int(FieldTextLength)
	FieldLink := self.FieldLink
	FieldLinkArg, FieldLinkFin := WrapString(FieldLink)
	result.link = FieldLinkArg
	FieldLinkLength := self.FieldLinkLength

	result.linkLength = C.int(FieldLinkLength)
	FieldUserData := self.FieldUserData
	FieldUserDataArg, FieldUserDataFin := WrapVoidPtr(FieldUserData)
	result.userData = FieldUserDataArg
	FieldIsImage := self.FieldIsImage

	result.isImage = C.bool(FieldIsImage)
	releaseFn = func() {
		FieldTextFin()

		FieldLinkFin()

		FieldUserDataFin()
	}
	return result, releaseFn
}

func (self MarkdownLinkCallbackData) c() (result C.MarkdownLinkCallbackData, fin func()) {
	resultPtr, finFn := self.handle()
	return *resultPtr, finFn
}

func newMarkdownLinkCallbackDataFromC(cvalue *C.MarkdownLinkCallbackData) *MarkdownLinkCallbackData {
	result := new(MarkdownLinkCallbackData)
	result.FieldText = C.GoString(cvalue.text)
	result.FieldTextLength = int32(cvalue.textLength)
	result.FieldLink = C.GoString(cvalue.link)
	result.FieldLinkLength = int32(cvalue.linkLength)
	result.FieldUserData = unsafe.Pointer(cvalue.userData)
	result.FieldIsImage = cvalue.isImage == C.bool(true)
	return result
}

type MarkdownTooltipCallbackData struct {
	FieldLinkData MarkdownLinkCallbackData
	FieldLinkIcon string
}

func (self MarkdownTooltipCallbackData) handle() (result *C.MarkdownTooltipCallbackData, releaseFn func()) {
	result = new(C.MarkdownTooltipCallbackData)
	FieldLinkData := self.FieldLinkData
	FieldLinkDataArg, FieldLinkDataFin := FieldLinkData.c()
	result.linkData = FieldLinkDataArg
	FieldLinkIcon := self.FieldLinkIcon
	FieldLinkIconArg, FieldLinkIconFin := WrapString(FieldLinkIcon)
	result.linkIcon = FieldLinkIconArg
	releaseFn = func() {
		FieldLinkDataFin()
		FieldLinkIconFin()
	}
	return result, releaseFn
}

func (self MarkdownTooltipCallbackData) c() (result C.MarkdownTooltipCallbackData, fin func()) {
	resultPtr, finFn := self.handle()
	return *resultPtr, finFn
}

func newMarkdownTooltipCallbackDataFromC(cvalue *C.MarkdownTooltipCallbackData) *MarkdownTooltipCallbackData {
	result := new(MarkdownTooltipCallbackData)
	result.FieldLinkData = *newMarkdownLinkCallbackDataFromC(func() *C.MarkdownLinkCallbackData { result := cvalue.linkData; return &result }())

	result.FieldLinkIcon = C.GoString(cvalue.linkIcon)
	return result
}

type TextBlock struct {
	FieldStart int32
	FieldStop  int32
}

func (self TextBlock) handle() (result *C.TextBlock, releaseFn func()) {
	result = new(C.TextBlock)
	FieldStart := self.FieldStart

	result.start = C.int(FieldStart)
	FieldStop := self.FieldStop

	result.stop = C.int(FieldStop)
	releaseFn = func() {
	}
	return result, releaseFn
}

func (self TextBlock) c() (result C.TextBlock, fin func()) {
	resultPtr, finFn := self.handle()
	return *resultPtr, finFn
}

func newTextBlockFromC(cvalue *C.TextBlock) *TextBlock {
	result := new(TextBlock)
	result.FieldStart = int32(cvalue.start)
	result.FieldStop = int32(cvalue.stop)
	return result
}

type TextRegion struct{}

func (self TextRegion) handle() (result *C.TextRegion, releaseFn func()) {
	result = new(C.TextRegion)
	releaseFn = func() {
	}
	return result, releaseFn
}

func (self TextRegion) c() (result C.TextRegion, fin func()) {
	resultPtr, finFn := self.handle()
	return *resultPtr, finFn
}

func newTextRegionFromC(cvalue *C.TextRegion) *TextRegion {
	result := new(TextRegion)
	return result
}
