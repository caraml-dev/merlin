FROM python:3.10

WORKDIR /mlobs
COPY sdk ./sdk
WORKDIR /mlobs/observation-publisher
COPY observation-publisher/requirements.txt .
RUN pip install -r requirements.txt
RUN rm requirements.txt
WORKDIR /mlobs
COPY observation-publisher ./observation-publisher
WORKDIR /mlobs/observation-publisher
ENTRYPOINT ["python", "-m", "publisher"]