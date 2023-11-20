FROM python:3.10

WORKDIR /mlobs

COPY observation-publisher/requirements.txt .
COPY sdk/ ./sdk
ENV SDK_PATH=/mlobs/sdk
RUN pip install -r requirements.txt
RUN rm requirements.txt
RUN rm -rf sdk
COPY observation-publisher/conf/ ./conf
COPY observation-publisher/publisher/ ./publisher

ENTRYPOINT ["python", "-m", "publisher"]