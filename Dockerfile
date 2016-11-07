# workspace (GOPATH) configured at /go.
FROM golang

RUN printf "machine bitbucket.org\nlogin tekion_build\npassword t3k10n_team" > ~/.netrc ## build account on bitbucket

# Copy the local package files to the container's workspace.
RUN mkdir -p /go/src/bitbucket.org/tekion
WORKDIR /go/src/bitbucket.org/tekion/tmessenger
#RUN git clone https://bitbucket.org/tekion/tmessenger.git
#WORKDIR /go/src/bitbucket.org/tekion/tmessenger
COPY . /go/src/bitbucket.org/tekion/tmessenger
#ADD . /go/src/bitbucket.org/tekion/tmessenger

# Build the tmessenger command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
#RUN go install bitbucket.org/tekion/tmessenger

RUN go-wrapper download

## From: https://medium.com/developers-writing/docker-powered-development-environment-for-your-go-app-6185d043ea35#.5093g1l8i
RUN go-wrapper install

# Set an environment variable
ENV MONGOSERVERS=MONGOSERVER:27017

# Run the tmessenger command by default when the container starts.
ENTRYPOINT /go/bin/tmessenger

# Document that the service listens on port 8080.
EXPOSE 8091
