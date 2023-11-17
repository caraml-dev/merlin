FROM python:3.10

WORKDIR /mlobs

COPY requirements.txt .
RUN pip install -r requirements.txt
RUN rm requirements.txt
COPY conf/ ./conf
COPY publisher/ ./publisher

ENTRYPOINT ["python", "-m", "publisher"]