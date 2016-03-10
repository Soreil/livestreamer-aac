package main

// #cgo pkg-config: libavcodec libavutil libavformat libswscale
/*

#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>
#include <libavutil/pixdesc.h>
#include <libavformat/avformat.h>
#include <libswscale/swscale.h>
#include <stdio.h>

#define BUFFER_SIZE 4096

struct buffer_data {
	uint8_t *ptr;
	size_t size; ///< size left in the buffer
};
static int read_packet(void *opaque, uint8_t *buf, int buf_size)
{
	struct buffer_data *bd = (struct buffer_data *)opaque;
	buf_size = FFMIN(buf_size, bd->size);
	// copy internal buffer data to buf
	memcpy(buf, bd->ptr, buf_size);
	bd->ptr  += buf_size;
	bd->size -= buf_size;
	return buf_size;
}


AVCodecContext * newAACEncoder(unsigned char *opaque,size_t len)
{

	unsigned char *buffer = (unsigned char*)av_malloc(BUFFER_SIZE+FF_INPUT_BUFFER_PADDING_SIZE);

	struct buffer_data bd = {0};
	bd.ptr = opaque;
	bd.size = len;

	//Allocate avioContext
	AVIOContext *ioCtx = avio_alloc_context(buffer,BUFFER_SIZE,0,&bd,&read_packet,NULL,NULL);

	AVFormatContext * ctx = avformat_alloc_context();

	//Set up context to read from memory
	ctx->pb = ioCtx;

	//open takes a fake filename when the context pb field is set up
	int err = avformat_open_input(&ctx, "dummyFileName", NULL, NULL);
	if (err < 0) {
		return NULL;
	}

	err = avformat_find_stream_info(ctx,NULL);
	if (err < 0) {
		return NULL;
	}

	AVCodec * codec = NULL;
	int strm = av_find_best_stream(ctx, AVMEDIA_TYPE_VIDEO, -1, -1, &codec, 0);

	AVCodecContext * codecCtx = ctx->streams[strm]->codec;
	return codecCtx;
}

	//	{
	//	err = avcodec_open2(codecCtx, codec, NULL);
	//	if (err < 0) {
	//		return NULL;
	//	}
	//
	//
	//	for (;;)
	//	{
	//		AVPacket pkt;
	//		err = av_read_frame(ctx, &pkt);
	//		if (err < 0) {
	//			return NULL;
	//		}
	//
	//		if (pkt.stream_index == strm)
	//		{
	//			int got = 0;
	//			AVFrame * frame = av_frame_alloc();
	//			err = avcodec_decode_video2(codecCtx, frame, &got, &pkt);
	//			if (err < 0) {
	//				return NULL;
	//			}
	//
	//			if (got)
	//			{
	//				//Throwing out the old stuff
	//				av_free(ioCtx);
	//				av_free(buffer);
	//				//avformat_free_context(ctx);
	//
	//				return frame;
	//			}
	//			av_frame_free(&frame);
	//		}
	//	}
	//}
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

func init() {
	C.av_register_all()
	C.avcodec_register_all()
}

func decode(data []byte) ([]byte, error) {
	enc := C.newAACEncoder((*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data)))
	if f == nil {
		return nil, errors.New("Failed to decode")
	}
	fmt.Println(C.GoString(C.av_get_pix_fmt_name(int32(f.format))))
}
