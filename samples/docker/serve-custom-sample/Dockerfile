FROM happy365/fast-style-transfer-build as builder
FROM tensorflow/tensorflow:2.16.2-gpu
RUN mv /etc/apt/sources.list.d /etc/apt/sources.list.d.bak && \
	apt-get update
RUN apt-get install apt-transport-https -y && \
	apt-get install git wget -y && \
	mv /etc/apt/sources.list.d.bak /etc/apt/sources.list.d
RUN pip install Pillow==3.4.2  && \
	pip install scipy==0.18.1  && \
	pip install numpy==1.11.2 && \
	pip install flask && \
	cd /root && \
	git clone https://github.com/floydhub/fast-style-transfer.git && \
	add-apt-repository ppa:jonathonf/ffmpeg-3 -y && \
	apt update && apt install -y ffmpeg libav-tools x264 x265
COPY --from=builder /input /input
RUN cp -a /root/fast-style-transfer/examples/style /output
WORKDIR /root/fast-style-transfer
