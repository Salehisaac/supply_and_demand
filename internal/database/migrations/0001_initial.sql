-- MySQL dump 10.13  Distrib 8.0.36, for Linux (x86_64)
--
-- Host: localhost    Database: supply_and_demand
-- ------------------------------------------------------
-- Server version	8.0.36-0ubuntu0.23.10.1

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `bot_settings`
--

DROP TABLE IF EXISTS `bot_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `bot_settings` (
  `id` int NOT NULL AUTO_INCREMENT,
  `bot_token` text NOT NULL,
  `is_active` tinyint(1) DEFAULT '1',
  `inactive_message` text,
  `main_admin_id` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `main_admin_id` (`main_admin_id`),
  CONSTRAINT `bot_settings_ibfk_1` FOREIGN KEY (`main_admin_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `bot_settings`
--

LOCK TABLES `bot_settings` WRITE;
/*!40000 ALTER TABLE `bot_settings` DISABLE KEYS */;
/*!40000 ALTER TABLE `bot_settings` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `categories`
--

DROP TABLE IF EXISTS `categories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `categories` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT '1',
  `inactive_message` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `categories`
--

LOCK TABLES `categories` WRITE;
/*!40000 ALTER TABLE `categories` DISABLE KEYS */;
INSERT INTO `categories` VALUES (1,'Sample Category','2024-05-21 22:41:37','2024-05-23 00:24:16',NULL,1,'Category is active'),(2,'Sample_Category 2','2024-05-21 22:52:41','2024-05-21 22:52:41',NULL,1,'Category is active');
/*!40000 ALTER TABLE `categories` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `categories_users`
--

DROP TABLE IF EXISTS `categories_users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `categories_users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `user_id` int DEFAULT NULL,
  `category_id` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  KEY `category_id` (`category_id`),
  CONSTRAINT `categories_users_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `categories_users_ibfk_2` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `categories_users`
--

LOCK TABLES `categories_users` WRITE;
/*!40000 ALTER TABLE `categories_users` DISABLE KEYS */;
/*!40000 ALTER TABLE `categories_users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `requests`
--

DROP TABLE IF EXISTS `requests`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `requests` (
  `id` int NOT NULL AUTO_INCREMENT,
  `responser_id` int DEFAULT NULL,
  `customer_id` int DEFAULT NULL,
  `category_id` int DEFAULT NULL,
  `text` text,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `tracking_code` varchar(255) DEFAULT NULL,
  `status` enum('پاسخ داده شده','پاسخ داده نشده','لغو شده') DEFAULT NULL,
  `request_id` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `responser_id` (`responser_id`),
  KEY `customer_id` (`customer_id`),
  KEY `category_id` (`category_id`),
  KEY `request_id` (`request_id`),
  CONSTRAINT `requests_ibfk_1` FOREIGN KEY (`responser_id`) REFERENCES `users` (`id`),
  CONSTRAINT `requests_ibfk_2` FOREIGN KEY (`customer_id`) REFERENCES `users` (`id`),
  CONSTRAINT `requests_ibfk_3` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`),
  CONSTRAINT `requests_ibfk_4` FOREIGN KEY (`request_id`) REFERENCES `requests` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=28 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `requests`
--

LOCK TABLES `requests` WRITE;
/*!40000 ALTER TABLE `requests` DISABLE KEYS */;
INSERT INTO `requests` VALUES (3,1,1,2,'salamkhoobi ?','2024-05-23 00:22:00',NULL,'HhdCXmPTrLzS',NULL,NULL),(4,1,1,1,' سشمشپ\n داداش گلم\n چه خبرا\n','2024-05-23 00:24:42',NULL,'NqVQEZnuzSzn',NULL,NULL),(6,2,2,1,' سلام\n حالت خوبه ؟\n احوالت خوبه ؟\n','2024-05-23 11:22:27',NULL,'dgJozK7BnHQv','پاسخ داده نشده',3),(7,2,2,1,' سلام\n حالتون خوبه ؟\n','2024-05-23 11:25:09',NULL,'wPDo6ievExE0','پاسخ داده نشده',3),(8,2,2,1,' سلام\n','2024-05-23 11:26:36',NULL,'XOXyJKPveMgh','پاسخ داده نشده',3),(9,2,2,2,' سلام\n','2024-05-23 11:28:57',NULL,'XuBJ1rC8XGes','پاسخ داده نشده',3),(10,2,2,1,' سلام\n','2024-05-23 11:30:09',NULL,'ZBoERgSCRpJP','پاسخ داده نشده',3),(11,2,2,1,'','2024-05-23 11:31:06',NULL,'A0OrqIKBt0Ua','پاسخ داده نشده',3),(12,2,2,1,' سلام\n','2024-05-23 11:31:20',NULL,'S5rTKMakkxbr','پاسخ داده نشده',3),(13,2,2,1,' سلام\n','2024-05-23 11:35:38',NULL,'Bu3hq0ArJnuZ','پاسخ داده نشده',3),(14,2,2,1,' سلام\n','2024-05-23 11:36:45',NULL,'08nuIxC9lrlw','پاسخ داده نشده',3),(15,2,2,1,' سلام\n','2024-05-23 11:37:47',NULL,'218jJPDOQCUP','پاسخ داده نشده',3),(16,2,2,1,' سلام بر تو\n','2024-05-23 11:39:59',NULL,'lhkjkzx9BgnZ','پاسخ داده نشده',3),(17,2,2,2,' salam\n','2024-05-23 11:56:25',NULL,'VIK1AYnK4bRz','پاسخ داده نشده',3),(18,2,2,2,' salam\n salam\n','2024-05-23 11:56:51',NULL,'Wfqd1tW8xx5Q','پاسخ داده نشده',3),(19,2,2,2,' salam\n salam\n salam in yek teste\n ke bratoon mifrestam\n','2024-05-23 12:00:11',NULL,'xIWBHGv6ywfB','پاسخ داده نشده',3),(20,2,2,1,' salam\n salam\n salam in yek teste\n ke bratoon mifrestam\n salam in ham yek teset ke mifrestam\n','2024-05-23 12:01:24',NULL,'4UM3Uzapk2Iz','پاسخ داده نشده',3),(21,2,2,1,' salam\n','2024-05-23 12:04:37',NULL,'nICQautIfsWy','پاسخ داده نشده',3),(22,2,2,2,' salm2\n','2024-05-23 12:06:49',NULL,'9ZTRMcngpdXj','پاسخ داده نشده',3),(23,2,2,1,' salam\n khoobi ?\n','2024-05-23 12:26:41',NULL,'VAmh87ooLdVW','پاسخ داده نشده',3),(24,2,2,1,' salam olagh aziz halet chetore\n','2024-05-23 20:13:10',NULL,'b3tgv2kmeeOo','پاسخ داده نشده',3),(25,2,2,1,' hi\n','2024-05-23 20:56:56',NULL,'v8YJMjrtvfA5','پاسخ داده نشده',3),(26,2,2,1,' hi\n','2024-05-23 21:02:09',NULL,'iubLPdX3bQVz','پاسخ داده نشده',3),(27,2,2,1,' hi\n','2024-05-23 21:51:52',NULL,'XEUw2GTMwtLY','پاسخ داده نشده',3);
/*!40000 ALTER TABLE `requests` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `chatId` bigint DEFAULT NULL,
  `username` varchar(120) DEFAULT NULL,
  `name` varchar(120) NOT NULL,
  `type` enum('admin','responder','customer') DEFAULT 'customer',
  `is_active` tinyint(1) DEFAULT '1',
  `inactive_message` text,
  `command` varchar(120) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `chatId` (`chatId`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (1,1234679,'saleh','YASHUA','admin',1,'','','2024-05-20 19:12:57'),(2,291109889,'salltin','','customer',1,'','','2024-05-20 20:41:07');
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2024-05-24  1:38:53