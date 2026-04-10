FROM docker-public.lwork.com:5000/base-jdk:v1.2
ENV artifact bw-user
ENV cfile app.toml
ENV workingDir /home/brokerwork
ENV user brokerwork

RUN useradd $user -m -d $workingDir && \
    mkdir $workingDir/log && \
        ln -s /dev/stdout $workingDir/log/process.log && \
    ln -s /dev/stdout $workingDir/log/access.log && \
    chown $user:$user $workingDir/log && \
    chmod 755 $workingDir 
ADD $artifact /home/$user/$artifact
ADD start-s.sh /home/$user/start.sh
COPY $cfile /home/$user/
COPY release.version /home/$user/
WORKDIR $workingDir 
RUN chmod +x $artifact start.sh
USER $user
CMD ["sh", "-c", "./start.sh ${artifact}"]
