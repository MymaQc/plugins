namespace Dragonfly;

public static class Net
{
    public sealed class UDPAddr
    {
        public UDPAddr(byte[] ip, int port, string zone = "")
        {
            ArgumentNullException.ThrowIfNull(ip);
            ArgumentNullException.ThrowIfNull(zone);
            IP = ip;
            Port = port;
            Zone = zone;
        }

        public byte[] IP { get; set; }
        public int Port { get; set; }
        public string Zone { get; set; }
    }
}
