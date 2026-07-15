namespace Dragonfly;

public static class Net
{
    public interface Addr
    {
        string Network();
        string String();
    }

    public sealed class UDPAddr : Addr
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

        public string Network() => "udp";

        public string String()
        {
            var mapped = IP.Length == 16 && IP.AsSpan(0, 10).SequenceEqual(new byte[10]) &&
                IP[10] == 0xff && IP[11] == 0xff;
            var bytes = mapped ? IP.AsSpan(12, 4).ToArray() : IP;
            var ip = new System.Net.IPAddress(bytes).ToString();
            if (Zone.Length != 0) ip += "%" + Zone;
            return ip.Contains(':') ? $"[{ip}]:{Port}" : $"{ip}:{Port}";
        }

        public override string ToString() => String();
    }

    internal sealed class AddrSnapshot(string network, string address) : Addr
    {
        public string Network() => network;
        public string String() => address;
        public override string ToString() => address;
    }
}
