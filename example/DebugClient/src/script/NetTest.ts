import Event = Laya.Event;
import Socket = Laya.Socket;
import Byte = Laya.Byte;
import { ui } from "../ui/layaMaxUI";

export default class NetTest extends ui.test.mainUI{
    private socket:Socket;
    private output:Byte;
    
    constructor() { 
        super(); 
        this.btn_ConnectServer.on(Event.CLICK,this,this.connectToServer);
        this.btn_CloseConnect.on(Event.CLICK,this,this.Close);
        this.btn_Send.on(Event.CLICK,this,this.sendMessage);
        
        this.btn_Login.on(Event.CLICK, this, function() {
            this.messageType.text ="2";
            this.onChangeInput('{"Account":"zllang1"}');
        });
        this.btn_Room.on(Event.CLICK, this, function() {
            this.messageType.text ="2";
            this.onChangeInput('{"UID":123}');
        });
    }
    
    //连接服务器
    connectToServer():void{
        this.socket = new Socket();
        this.socket.connectByUrl(this.addr.text);

        this.output = this.socket.output;

        this.socket.on(Event.OPEN, this, this.onSocketOpen);
        this.socket.on(Event.CLOSE, this, this.onSocketClose);
        this.socket.on(Event.MESSAGE, this, this.onMessageReveived);
        this.socket.on(Event.ERROR, this, this.onConnectError);

        
    }

    private onSocketOpen(): void {
        console.log("Connected");
        this.Log("连接服务器成功");
    }

    private sendMessage():void{
        // 使用output.writeByte发送
        var message: string = this.content.text;
        if (message == ""){
            alert("message can not be empty");
        }
        var mid :number = parseInt(this.messageType.text);
        var by:Byte=new Byte();
        by.endian = Byte.BIG_ENDIAN;

        by.writeUint32(mid);
        for (var i: number = 0; i < message.length; ++i) {
            by.writeByte(message.charCodeAt(i));
        }

        this.socket.send(by.buffer);
    }

    private onMessageReveived(message: any): void {
        console.log("Message from server:");
        if (typeof message == "string") {
            this.Log(message);
        }
        else if (message instanceof ArrayBuffer) {
            var by:Byte=new Byte(message);
            var messageID:number = by.readUint32();
            var str:string = by.readUTFString();
            this.Log('MessageID:${messageID} Message:${str}')
        }
        this.socket.input.clear();
    }

    private onChangeInput(str:string){
        this.content.text =str;
    }

    private Close():void{
        this.socket.close();
        this.Log("连接关闭");
    }
    private onSocketClose(): void {
        this.Log("连接关闭");
    }
    private onConnectError(e: Event): void {
        this.Log("连接错误");
    }

    private Log(str:string){
        var time:string=new Date().toTimeString();
        this.log.text+=(time+": \n"+str+"\n");
        this.log.scrollTo(this.log.maxScrollY);
    }

    onEnable(): void {
        
    }

    onDisable(): void {
    }
}