Passo a passo:
1. Abre o ficheiro com um editor: sudo nano /etc/fstab
2. Usa as setas do teclado para ir até ao fim do ficheiro
3. Adiciona esta linha no final: nome_da_pasta    /mnt/shared    vboxsf    defaults,uid=1000,gid=1000    0    0
Substitui nome_da_pasta pelo nome que deste à pasta partilhada no VirtualBox!

4. Guarda e sai:
Pressiona Ctrl + O (para gravar)
Pressiona Enter (para confirmar)
Pressiona Ctrl + X (para sair)


-------------------------------------
export GOCACHE=$HOME/.cache/go-build
mkdir -p $HOME/.cache/go-build
export PATH=$PATH:$GOPATH/bin