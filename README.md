# traffic-light-demo

Traffic light project deployment using Kubernetes and Kubeedge. Steps are down below:

First step:  

  Make the right pin mapping with your raspberry pi (raspberry pi 3 used for this demo), which is the following:  

  Red Led = gpio 9  

  Yellow Led = gpio 10  

  Green Led = gpio 11

Second step:  

  Apply manifest files in /crds  

  kubectl create -f devicemodel.yaml  

  kubectl create -f deviceinstance.yaml  

# Note: Don't forget to edit edge node name according to what you have set  

Third step:  

  Apply manifest file to run the container
  
  kubectl create -f deployment.yaml

# Traffic LEDs should be working. You can get device status from cloud using this command kubectl get device device-instance-name -o yaml. Real time values are found in reported field

