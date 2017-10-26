#!/bin/bash
apt-get update
apt-get install -fqqy git mercurial subversion curl vim-nox xvfb
mkdir /usr/local/src/download
cd /usr/local/src/download

echo "Download Go"
wget --quiet https://storage.googleapis.com/golang/go1.9.1.linux-amd64.tar.gz

echo "Download Chrome"
wget --quiet http://mirror.cs.uchicago.edu/google-chrome/pool/main/g/google-chrome-stable/google-chrome-stable_61.0.3163.100-1_amd64.deb

echo "Install Go"
cd /usr/local/
tar -xf /usr/local/src/download/go1.9.1.linux-amd64.tar.gz

cp /vagrant/profile.d-go.sh /etc/profile.d/go.sh
chmod a+rx /etc/profile.d/go.sh

cp /vagrant/profile.d-cshot-server.sh /etc/profile.d/cshot-server.sh
chmod a+rx /etc/profile.d/cshot-server.sh

echo "Install Chrome"
apt -fqqy install /usr/local/src/download/google-chrome-stable_61.0.3163.100-1_amd64.deb

echo "Build cshot-server"
source /etc/profile.d/go.sh
go get github.com/apokalyptik/cshot/cmd/cshot-server
go install github.com/apokalyptik/cshot/cmd/cshot-server

echo "Start cshot-server"
cp -v /vagrant/cshot-server.service /lib/systemd/system/cshot-server.service
systemctl daemon-reload
systemctl enable cshot-server.service
systemctl start cshot-server.service

echo "Setup Vim"
curl --silent https://gist.githubusercontent.com/apokalyptik/fdf050e2dd004b756d2e1a0b6f2d399a/raw/.vimrc > /root/.vimrc
git clone https://github.com/VundleVim/Vundle.vim.git /root/.vim/bundle/Vundle.vim
echo | vim +PluginInstall +qall > /dev/null
vim +GoInstallBinaries +qall > /dev/null

echo "The above warnings about input and output from vim are expected. And this echo is here to exit with a success as well as inform"
echo
echo
echo "With any luck the service should be up and running. visit http://127.0.0.1:8001/cshot/v1/http:/google.com"
